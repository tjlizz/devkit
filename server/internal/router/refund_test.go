package router

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

// TestRefundFlow verifies a buyer can refund their own paid order, which
// revokes the entitlement, is idempotent, and drops the order from the
// developer's sales revenue.
func TestRefundFlow(t *testing.T) {
	db, app := newAuthTestApp(t)
	devID := insertVerifiedUser(t, db, "dev@example.com", "old-password")
	buyerID := insertVerifiedUser(t, db, "buyer@example.com", "old-password")
	otherID := insertVerifiedUser(t, db, "other@example.com", "old-password")
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

	// Publish and approve an app, then buy it.
	appID := publishAndApprove(t, app, developerToken, adminToken, "Refund App", "refund-app")
	checkout := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/marketplace/apps/refund-app/checkout",
		"{}",
		buyerToken,
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
	orderID := checkoutBody.Order.ID

	confirm := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/orders/"+strconv.FormatInt(orderID, 10)+"/confirm-payment",
		"",
		buyerToken,
	)
	if confirm.Code != http.StatusOK {
		t.Fatalf("confirm status = %d, want %d; body = %s", confirm.Code, http.StatusOK, confirm.Body.String())
	}
	var confirmBody struct {
		Entitlement struct {
			ID int64 `json:"id"`
		} `json:"entitlement"`
	}
	if err := json.NewDecoder(confirm.Body).Decode(&confirmBody); err != nil {
		t.Fatalf("decode confirm: %v", err)
	}
	entID := confirmBody.Entitlement.ID

	// --- Unauthenticated refund is rejected. ---
	unauth := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/orders/"+strconv.FormatInt(orderID, 10)+"/refund",
		"",
	)
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated refund status = %d, want %d", unauth.Code, http.StatusUnauthorized)
	}

	// --- A different user cannot refund the buyer's order. ---
	foreign := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/orders/"+strconv.FormatInt(orderID, 10)+"/refund",
		"",
		otherToken,
	)
	if foreign.Code != http.StatusForbidden {
		t.Fatalf("foreign refund status = %d, want %d", foreign.Code, http.StatusForbidden)
	}

	// --- Refund flips the order and revokes the entitlement. ---
	refund := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/orders/"+strconv.FormatInt(orderID, 10)+"/refund",
		"",
		buyerToken,
	)
	if refund.Code != http.StatusOK {
		t.Fatalf("refund status = %d, want %d; body = %s", refund.Code, http.StatusOK, refund.Body.String())
	}
	var refundBody struct {
		Order struct {
			Status string `json:"status"`
		} `json:"order"`
	}
	if err := json.NewDecoder(refund.Body).Decode(&refundBody); err != nil {
		t.Fatalf("decode refund: %v", err)
	}
	if refundBody.Order.Status != "refunded" {
		t.Fatalf("refunded order status = %q, want %q", refundBody.Order.Status, "refunded")
	}

	var entStatus string
	if err := db.QueryRow("SELECT status FROM entitlements WHERE id = ?", entID).Scan(&entStatus); err != nil {
		t.Fatalf("read entitlement status: %v", err)
	}
	if entStatus != "revoked" {
		t.Fatalf("entitlement status = %q, want %q", entStatus, "revoked")
	}

	// --- Refunding again is idempotent. ---
	refundAgain := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/orders/"+strconv.FormatInt(orderID, 10)+"/refund",
		"",
		buyerToken,
	)
	if refundAgain.Code != http.StatusOK {
		t.Fatalf("re-refund status = %d, want %d", refundAgain.Code, http.StatusOK)
	}

	// --- Refunded order drops out of the developer's sales revenue. ---
	sales := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/developer/sales", "", developerToken)
	if sales.Code != http.StatusOK {
		t.Fatalf("sales status = %d, want %d", sales.Code, http.StatusOK)
	}
	var salesBody struct {
		Sales   []json.RawMessage `json:"sales"`
		Summary struct {
			TotalOrders       int64 `json:"totalOrders"`
			TotalRevenueCents int64 `json:"totalRevenueCents"`
		} `json:"summary"`
	}
	if err := json.NewDecoder(sales.Body).Decode(&salesBody); err != nil {
		t.Fatalf("decode sales: %v", err)
	}
	if len(salesBody.Sales) != 0 || salesBody.Summary.TotalOrders != 0 || salesBody.Summary.TotalRevenueCents != 0 {
		t.Fatalf("sales after refund = %+v, want empty", salesBody)
	}

	// --- The revoked entitlement blocks delivery access. ---
	delivery := performJSONRequestWithToken(
		t,
		app,
		http.MethodGet,
		"/api/v1/entitlements/"+strconv.FormatInt(entID, 10)+"/delivery",
		"",
		buyerToken,
	)
	if delivery.Code != http.StatusForbidden {
		t.Fatalf("delivery for revoked entitlement status = %d, want %d", delivery.Code, http.StatusForbidden)
	}

	// Sanity references.
	if buyerID < 1 || otherID < 1 || appID < 1 {
		t.Fatalf("unexpected ids buyer=%d other=%d app=%d", buyerID, otherID, appID)
	}
}

// TestRefundRejectsPendingOrder verifies a pending (unpaid) order cannot be
// refunded.
func TestRefundRejectsPendingOrder(t *testing.T) {
	db, app := newAuthTestApp(t)
	devID := insertVerifiedUser(t, db, "dev@example.com", "old-password")
	_ = insertVerifiedUser(t, db, "buyer@example.com", "old-password")
	adminID := insertVerifiedUser(t, db, "admin@example.com", "old-password")
	if _, err := db.Exec("UPDATE users SET role = 'admin' WHERE id = ?", adminID); err != nil {
		t.Fatalf("mark admin: %v", err)
	}
	if _, err := db.Exec("INSERT INTO developers(user_id, display_name) VALUES (?, ?)", devID, "Dev Studio"); err != nil {
		t.Fatalf("insert developer: %v", err)
	}

	developerToken := loginToken(t, app, "dev@example.com", "old-password")
	buyerToken := loginToken(t, app, "buyer@example.com", "old-password")
	adminToken := loginToken(t, app, "admin@example.com", "old-password")

	_ = publishAndApprove(t, app, developerToken, adminToken, "Pending App", "pending-app")
	checkout := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/marketplace/apps/pending-app/checkout",
		"{}",
		buyerToken,
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

	refund := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/orders/"+strconv.FormatInt(checkoutBody.Order.ID, 10)+"/refund",
		"",
		buyerToken,
	)
	if refund.Code != http.StatusConflict {
		t.Fatalf("refund pending status = %d, want %d; body = %s", refund.Code, http.StatusConflict, refund.Body.String())
	}
}
