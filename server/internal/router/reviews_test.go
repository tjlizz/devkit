package router

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

// TestReviewFlow verifies that only verified buyers with an active entitlement
// can create/update a review, that each buyer has at most one review per app,
// that reviews are publicly readable, that aggregates are returned on the app,
// and that a buyer can delete their own review.
func TestReviewFlow(t *testing.T) {
	db, app := newAuthTestApp(t)
	devID := insertVerifiedUser(t, db, "dev@example.com", "old-password")
	insertVerifiedUser(t, db, "buyer@example.com", "old-password")
	insertVerifiedUser(t, db, "other@example.com", "old-password")
	adminID := insertVerifiedUser(t, db, "admin@example.com", "old-password")
	if _, err := db.Exec("UPDATE users SET role = 'admin' WHERE id = ?", adminID); err != nil {
		t.Fatalf("mark admin: %v", err)
	}
	if _, err := db.Exec("INSERT INTO developers(user_id, display_name) VALUES (?, ?)", devID, "Dev Studio"); err != nil {
		t.Fatalf("insert developer: %v", err)
	}

	developerToken := loginToken(t, app, "dev@example.com", "old-password")
	buyerToken := loginToken(t, app, "buyer@example.com", "old-password")
	otherToken := loginToken(t, app, "other@example.com", "old-password")
	adminToken := loginToken(t, app, "admin@example.com", "old-password")

	publishAndApprove(t, app, developerToken, adminToken, "Reviewed App", "reviewed-app")

	// --- Unauthenticated create is rejected. ---
	unauth := performJSONRequest(
		t, app, http.MethodPost, "/api/v1/marketplace/apps/reviewed-app/reviews",
		`{"rating":5,"comment":"Great!"}`,
	)
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated review status = %d, want %d", unauth.Code, http.StatusUnauthorized)
	}

	// --- A buyer without an entitlement cannot review. ---
	noEntitlement := performJSONRequestWithToken(
		t, app, http.MethodPost, "/api/v1/marketplace/apps/reviewed-app/reviews",
		`{"rating":5,"comment":"Not a buyer."}`, otherToken,
	)
	if noEntitlement.Code != http.StatusForbidden {
		t.Fatalf("review without entitlement status = %d, want %d; body = %s",
			noEntitlement.Code, http.StatusForbidden, noEntitlement.Body.String())
	}

	// --- Invalid rating is rejected even for an entitled buyer. ---
	// Buyer purchases the app first.
	checkout := performJSONRequestWithToken(
		t, app, http.MethodPost, "/api/v1/marketplace/apps/reviewed-app/checkout", `{}`, buyerToken,
	)
	if checkout.Code != http.StatusCreated {
		t.Fatalf("checkout status = %d, want %d; body = %s", checkout.Code, http.StatusCreated, checkout.Body.String())
	}
	var checkoutBody struct {
		Order struct {
			ID int64 `json:"id"`
		} `json:"order"`
	}
	if err := json.NewDecoder(checkout.Body).Decode(&checkoutBody); err != nil {
		t.Fatalf("decode checkout: %v", err)
	}
	confirm := performJSONRequestWithToken(
		t, app, http.MethodPost,
		"/api/v1/orders/"+strconv.FormatInt(checkoutBody.Order.ID, 10)+"/confirm-payment",
		`{}`, buyerToken,
	)
	if confirm.Code != http.StatusOK {
		t.Fatalf("confirm payment status = %d, want %d; body = %s", confirm.Code, http.StatusOK, confirm.Body.String())
	}

	badRating := performJSONRequestWithToken(
		t, app, http.MethodPost, "/api/v1/marketplace/apps/reviewed-app/reviews",
		`{"rating":6,"comment":"Too high."}`, buyerToken,
	)
	if badRating.Code != http.StatusBadRequest {
		t.Fatalf("invalid rating status = %d, want %d", badRating.Code, http.StatusBadRequest)
	}

	// --- Entitled buyer creates a review. ---
	create := performJSONRequestWithToken(
		t, app, http.MethodPost, "/api/v1/marketplace/apps/reviewed-app/reviews",
		`{"rating":5,"comment":"Excellent tool!"}`, buyerToken,
	)
	if create.Code != http.StatusOK {
		t.Fatalf("create review status = %d, want %d; body = %s", create.Code, http.StatusOK, create.Body.String())
	}
	var created struct {
		Review struct {
			ID               int64  `json:"id"`
			Rating           int    `json:"rating"`
			Comment          string `json:"comment"`
			VerifiedPurchase bool   `json:"verifiedPurchase"`
		} `json:"review"`
	}
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode created review: %v", err)
	}
	if created.Review.Rating != 5 || created.Review.Comment != "Excellent tool!" {
		t.Fatalf("created review = %+v, want rating 5 / comment Excellent tool!", created.Review)
	}
	if !created.Review.VerifiedPurchase {
		t.Fatal("created review should be a verified purchase")
	}

	// --- Updating keeps one review per buyer (upsert). ---
	update := performJSONRequestWithToken(
		t, app, http.MethodPost, "/api/v1/marketplace/apps/reviewed-app/reviews",
		`{"rating":4,"comment":"Updated."}`, buyerToken,
	)
	if update.Code != http.StatusOK {
		t.Fatalf("update review status = %d, want %d; body = %s", update.Code, http.StatusOK, update.Body.String())
	}
	var updated struct {
		Review struct {
			ID      int64  `json:"id"`
			Rating  int    `json:"rating"`
			Comment string `json:"comment"`
		} `json:"review"`
	}
	if err := json.NewDecoder(update.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated review: %v", err)
	}
	if updated.Review.ID != created.Review.ID || updated.Review.Rating != 4 || updated.Review.Comment != "Updated." {
		t.Fatalf("updated review = %+v, want same id %d rating 4", updated.Review, created.Review.ID)
	}

	// --- App aggregates reflect the single review. ---
	appDetail := performJSONRequest(t, app, http.MethodGet, "/api/v1/marketplace/apps/reviewed-app", "")
	if appDetail.Code != http.StatusOK {
		t.Fatalf("app detail status = %d, want %d", appDetail.Code, http.StatusOK)
	}
	var appDetailBody struct {
		App struct {
			Rating      float64 `json:"rating"`
			ReviewCount int     `json:"reviewCount"`
		} `json:"app"`
	}
	if err := json.NewDecoder(appDetail.Body).Decode(&appDetailBody); err != nil {
		t.Fatalf("decode app detail: %v", err)
	}
	if appDetailBody.App.ReviewCount != 1 || appDetailBody.App.Rating != 4.0 {
		t.Fatalf("app aggregates = %+v, want count 1 rating 4.0", appDetailBody.App)
	}

	// --- Public list returns the review. ---
	list := performJSONRequest(t, app, http.MethodGet, "/api/v1/marketplace/apps/reviewed-app/reviews", "")
	if list.Code != http.StatusOK {
		t.Fatalf("list reviews status = %d, want %d", list.Code, http.StatusOK)
	}
	var listBody struct {
		Reviews []struct {
			ID       int64 `json:"id"`
			BuyerID  int64 `json:"buyerId"`
			Rating   int   `json:"rating"`
			Verified bool  `json:"verifiedPurchase"`
		} `json:"reviews"`
	}
	if err := json.NewDecoder(list.Body).Decode(&listBody); err != nil {
		t.Fatalf("decode list reviews: %v", err)
	}
	if len(listBody.Reviews) != 1 {
		t.Fatalf("list reviews length = %d, want 1", len(listBody.Reviews))
	}
	if listBody.Reviews[0].ID != created.Review.ID || listBody.Reviews[0].Rating != 4 {
		t.Fatalf("listed review = %+v, want id %d rating 4", listBody.Reviews[0], created.Review.ID)
	}

	// --- My review endpoint returns the buyer's review. ---
	mine := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/marketplace/apps/reviewed-app/reviews/me", "", buyerToken)
	if mine.Code != http.StatusOK {
		t.Fatalf("my review status = %d, want %d; body = %s", mine.Code, http.StatusOK, mine.Body.String())
	}

	// --- Buyer deletes their review; it disappears from the public list. ---
	del := performJSONRequestWithToken(t, app, http.MethodDelete, "/api/v1/marketplace/apps/reviewed-app/reviews/me", "", buyerToken)
	if del.Code != http.StatusOK {
		t.Fatalf("delete review status = %d, want %d; body = %s", del.Code, http.StatusOK, del.Body.String())
	}
	afterDelete := performJSONRequest(t, app, http.MethodGet, "/api/v1/marketplace/apps/reviewed-app/reviews", "")
	var afterBody struct {
		Reviews []json.RawMessage `json:"reviews"`
	}
	if err := json.NewDecoder(afterDelete.Body).Decode(&afterBody); err != nil {
		t.Fatalf("decode after delete: %v", err)
	}
	if len(afterBody.Reviews) != 0 {
		t.Fatalf("reviews after delete = %d, want 0", len(afterBody.Reviews))
	}

	// --- Aggregates reset to zero after deletion. ---
	appDetail2 := performJSONRequest(t, app, http.MethodGet, "/api/v1/marketplace/apps/reviewed-app", "")
	var appDetailBody2 struct {
		App struct {
			Rating      float64 `json:"rating"`
			ReviewCount int     `json:"reviewCount"`
		} `json:"app"`
	}
	if err := json.NewDecoder(appDetail2.Body).Decode(&appDetailBody2); err != nil {
		t.Fatalf("decode app detail after delete: %v", err)
	}
	if appDetailBody2.App.ReviewCount != 0 || appDetailBody2.App.Rating != 0 {
		t.Fatalf("aggregates after delete = %+v, want count 0 rating 0", appDetailBody2.App)
	}

	// --- A non-entitled user cannot see or have a "my review" (404). ---
	othersMine := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/marketplace/apps/reviewed-app/reviews/me", "", otherToken)
	if othersMine.Code != http.StatusNotFound {
		t.Fatalf("other user my-review status = %d, want %d", othersMine.Code, http.StatusNotFound)
	}
}
