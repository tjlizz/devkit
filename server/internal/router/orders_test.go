package router

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestOrderCheckoutConfirmAndEntitlementFlow(t *testing.T) {
	db, app := newAuthTestApp(t)
	devID := insertVerifiedUser(t, db, "dev@example.com", "old-password")
	buyerID := insertVerifiedUser(t, db, "buyer@example.com", "old-password")
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

	// Publish an app with one paid plan and one free plan.
	body := `{"name":"Build Lens","slug":"build-lens","tagline":"Release intelligence.","description":"Track release health.","category":"developer-tools","priceCents":4900,"currency":"USD","iconUrl":"https://example.com/icon.png","coverImageUrl":"","demoUrl":"https://example.com/demo","docsUrl":"https://example.com/docs","sourceUrl":"https://github.com/example/build-lens","supportUrl":"","tags":["release"],"version":"1.0.0","releaseNotes":"Initial release.","plans":[{"name":"Pro","priceCents":4900,"currency":"USD","description":"Pro plan","features":["Feature A"]},{"name":"Free","priceCents":0,"currency":"USD","description":"Free plan","features":[]}]}`
	submitted := performJSONRequestWithToken(t, app, http.MethodPost, "/api/v1/apps", body, developerToken)
	if submitted.Code != http.StatusCreated {
		t.Fatalf("publish status = %d, want %d; body = %s", submitted.Code, http.StatusCreated, submitted.Body.String())
	}
	var submittedBody struct {
		App struct {
			ID   int64  `json:"id"`
			Slug string `json:"slug"`
		} `json:"app"`
	}
	if err := json.NewDecoder(submitted.Body).Decode(&submittedBody); err != nil {
		t.Fatalf("decode published app: %v", err)
	}

	approved := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/admin/apps/"+strconv.FormatInt(submittedBody.App.ID, 10)+"/approve",
		`{"reviewNote":"Approved."}`,
		adminToken,
	)
	if approved.Code != http.StatusOK {
		t.Fatalf("approve app status = %d, want %d; body = %s", approved.Code, http.StatusOK, approved.Body.String())
	}

	// Read plan IDs from the marketplace detail response.
	detail := performJSONRequest(t, app, http.MethodGet, "/api/v1/marketplace/apps/build-lens", "")
	if detail.Code != http.StatusOK {
		t.Fatalf("marketplace detail status = %d, want %d", detail.Code, http.StatusOK)
	}
	var detailBody struct {
		App struct {
			Plans []struct {
				ID         int64  `json:"id"`
				Name       string `json:"name"`
				PriceCents int64  `json:"priceCents"`
			} `json:"plans"`
		} `json:"app"`
	}
	if err := json.NewDecoder(detail.Body).Decode(&detailBody); err != nil {
		t.Fatalf("decode marketplace detail: %v", err)
	}
	var proPlanID, freePlanID int64
	for _, plan := range detailBody.App.Plans {
		switch plan.Name {
		case "Pro":
			proPlanID = plan.ID
		case "Free":
			freePlanID = plan.ID
		}
	}
	if proPlanID == 0 || freePlanID == 0 {
		t.Fatalf("plans not found: %+v", detailBody.App.Plans)
	}

	// --- Checkout requires authentication. ---
	unauthenticated := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/marketplace/apps/build-lens/checkout",
		fmt.Sprintf(`{"planId":%d}`, proPlanID),
	)
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated checkout status = %d, want %d", unauthenticated.Code, http.StatusUnauthorized)
	}

	// --- Checkout creates a pending order with price snapshot. ---
	checkout := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/marketplace/apps/build-lens/checkout",
		fmt.Sprintf(`{"planId":%d}`, proPlanID),
		buyerToken,
	)
	if checkout.Code != http.StatusCreated {
		t.Fatalf("checkout status = %d, want %d; body = %s", checkout.Code, http.StatusCreated, checkout.Body.String())
	}
	var checkoutBody struct {
		Order struct {
			ID         int64  `json:"id"`
			Status     string `json:"status"`
			PlanName   string `json:"planName"`
			PriceCents int64  `json:"priceCents"`
			Currency   string `json:"currency"`
		} `json:"order"`
	}
	if err := json.NewDecoder(checkout.Body).Decode(&checkoutBody); err != nil {
		t.Fatalf("decode checkout: %v", err)
	}
	orderID := checkoutBody.Order.ID
	if orderID < 1 || checkoutBody.Order.Status != "pending" ||
		checkoutBody.Order.PlanName != "Pro" || checkoutBody.Order.PriceCents != 4900 ||
		checkoutBody.Order.Currency != "USD" {
		t.Fatalf("checkout order = %+v, want pending Pro 4900 USD", checkoutBody.Order)
	}

	// --- A different user cannot confirm the buyer's order. ---
	forbidden := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/orders/"+strconv.FormatInt(orderID, 10)+"/confirm-payment",
		"",
		otherToken,
	)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("foreign confirm status = %d, want %d", forbidden.Code, http.StatusForbidden)
	}

	// --- Confirm payment issues the entitlement. ---
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
		Order struct {
			Status          string `json:"status"`
			ProviderEventID string `json:"providerEventId"`
			PaidAt          string `json:"paidAt"`
		} `json:"order"`
		Entitlement struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"entitlement"`
	}
	if err := json.NewDecoder(confirm.Body).Decode(&confirmBody); err != nil {
		t.Fatalf("decode confirm: %v", err)
	}
	if confirmBody.Order.Status != "paid" || confirmBody.Order.ProviderEventID == "" || confirmBody.Order.PaidAt == "" {
		t.Fatalf("confirmed order = %+v, want paid with event id and paid_at", confirmBody.Order)
	}
	if confirmBody.Entitlement.ID < 1 || confirmBody.Entitlement.Status != "active" {
		t.Fatalf("entitlement = %+v, want active", confirmBody.Entitlement)
	}

	var entitlementCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM entitlements WHERE buyer_id = ? AND app_id = ?", buyerID, submittedBody.App.ID).Scan(&entitlementCount); err != nil {
		t.Fatalf("count entitlements: %v", err)
	}
	if entitlementCount != 1 {
		t.Fatalf("entitlement count = %d, want 1", entitlementCount)
	}

	// --- Confirming again is idempotent: no second entitlement, same result. ---
	reconfirm := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/orders/"+strconv.FormatInt(orderID, 10)+"/confirm-payment",
		"",
		buyerToken,
	)
	if reconfirm.Code != http.StatusOK {
		t.Fatalf("reconfirm status = %d, want %d", reconfirm.Code, http.StatusOK)
	}
	var reconfirmBody struct {
		Order struct {
			Status string `json:"status"`
		} `json:"order"`
	}
	if err := json.NewDecoder(reconfirm.Body).Decode(&reconfirmBody); err != nil {
		t.Fatalf("decode reconfirm: %v", err)
	}
	if reconfirmBody.Order.Status != "paid" {
		t.Fatalf("reconfirmed order status = %q, want paid", reconfirmBody.Order.Status)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM entitlements WHERE buyer_id = ? AND app_id = ?", buyerID, submittedBody.App.ID).Scan(&entitlementCount); err != nil {
		t.Fatalf("count entitlements after reconfirm: %v", err)
	}
	if entitlementCount != 1 {
		t.Fatalf("entitlement count after reconfirm = %d, want 1", entitlementCount)
	}

	// --- Duplicate purchase is rejected while entitlement is active. ---
	duplicate := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/marketplace/apps/build-lens/checkout",
		fmt.Sprintf(`{"planId":%d}`, freePlanID),
		buyerToken,
	)
	if duplicate.Code != http.StatusConflict {
		t.Fatalf("duplicate checkout status = %d, want %d", duplicate.Code, http.StatusConflict)
	}

	// --- My orders lists the paid order. ---
	myOrders := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/me/orders", "", buyerToken)
	if myOrders.Code != http.StatusOK {
		t.Fatalf("my orders status = %d, want %d", myOrders.Code, http.StatusOK)
	}
	var myOrdersBody struct {
		Orders []struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"orders"`
	}
	if err := json.NewDecoder(myOrders.Body).Decode(&myOrdersBody); err != nil {
		t.Fatalf("decode my orders: %v", err)
	}
	if len(myOrdersBody.Orders) != 1 || myOrdersBody.Orders[0].ID != orderID || myOrdersBody.Orders[0].Status != "paid" {
		t.Fatalf("my orders = %+v, want one paid order", myOrdersBody.Orders)
	}

	// --- My entitlements lists the active entitlement. ---
	myEnts := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/me/entitlements", "", buyerToken)
	if myEnts.Code != http.StatusOK {
		t.Fatalf("my entitlements status = %d, want %d", myEnts.Code, http.StatusOK)
	}
	var myEntsBody struct {
		Entitlements []struct {
			ID      int64  `json:"id"`
			Status  string `json:"status"`
			AppSlug string `json:"appSlug"`
		} `json:"entitlements"`
	}
	if err := json.NewDecoder(myEnts.Body).Decode(&myEntsBody); err != nil {
		t.Fatalf("decode my entitlements: %v", err)
	}
	if len(myEntsBody.Entitlements) != 1 || myEntsBody.Entitlements[0].Status != "active" ||
		myEntsBody.Entitlements[0].AppSlug != "build-lens" {
		t.Fatalf("my entitlements = %+v, want one active build-lens", myEntsBody.Entitlements)
	}
	entitlementID := myEntsBody.Entitlements[0].ID

	// --- Delivery is restricted to the owner and returns a signed token. ---
	foreignDelivery := performJSONRequestWithToken(
		t,
		app,
		http.MethodGet,
		"/api/v1/entitlements/"+strconv.FormatInt(entitlementID, 10)+"/delivery",
		"",
		otherToken,
	)
	if foreignDelivery.Code != http.StatusForbidden {
		t.Fatalf("foreign delivery status = %d, want %d", foreignDelivery.Code, http.StatusForbidden)
	}

	delivery := performJSONRequestWithToken(
		t,
		app,
		http.MethodGet,
		"/api/v1/entitlements/"+strconv.FormatInt(entitlementID, 10)+"/delivery",
		"",
		buyerToken,
	)
	if delivery.Code != http.StatusOK {
		t.Fatalf("delivery status = %d, want %d; body = %s", delivery.Code, http.StatusOK, delivery.Body.String())
	}
	var deliveryBody struct {
		EntitlementID int64  `json:"entitlementId"`
		AppSlug       string `json:"appSlug"`
		Version       string `json:"version"`
		SourceURL     string `json:"sourceUrl"`
		DeliveryToken string `json:"deliveryToken"`
		ExpiresAt     string `json:"expiresAt"`
	}
	if err := json.NewDecoder(delivery.Body).Decode(&deliveryBody); err != nil {
		t.Fatalf("decode delivery: %v", err)
	}
	if deliveryBody.EntitlementID != entitlementID || deliveryBody.AppSlug != "build-lens" ||
		deliveryBody.Version != "1.0.0" || deliveryBody.SourceURL == "" ||
		deliveryBody.DeliveryToken == "" {
		t.Fatalf("delivery = %+v, want full delivery payload", deliveryBody)
	}
	expiresAt, err := time.Parse(time.RFC3339, deliveryBody.ExpiresAt)
	if err != nil {
		t.Fatalf("parse expiresAt: %v", err)
	}
	if !expiresAt.After(time.Now()) {
		t.Fatalf("delivery token already expired: %s", deliveryBody.ExpiresAt)
	}

	// Verify the token is a signed payload of the entitlement ID and expiry.
	decoded, err := base64.RawURLEncoding.DecodeString(deliveryBody.DeliveryToken)
	if err != nil {
		t.Fatalf("decode delivery token: %v", err)
	}
	parts := strings.SplitN(string(decoded), ":", 3)
	if len(parts) != 3 {
		t.Fatalf("delivery token parts = %v, want 3", parts)
	}
	if parts[0] != strconv.FormatInt(entitlementID, 10) {
		t.Fatalf("delivery token entitlement = %s, want %d", parts[0], entitlementID)
	}

	// --- Other users have no orders or entitlements. ---
	otherOrders := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/me/orders", "", otherToken)
	var otherOrdersBody struct {
		Orders []struct {
			ID int64 `json:"id"`
		} `json:"orders"`
	}
	if err := json.NewDecoder(otherOrders.Body).Decode(&otherOrdersBody); err != nil {
		t.Fatalf("decode other orders: %v", err)
	}
	if len(otherOrdersBody.Orders) != 0 {
		t.Fatalf("other user orders = %+v, want none", otherOrdersBody.Orders)
	}

	otherEnts := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/me/entitlements", "", otherToken)
	var otherEntsBody struct {
		Entitlements []struct {
			ID int64 `json:"id"`
		} `json:"entitlements"`
	}
	if err := json.NewDecoder(otherEnts.Body).Decode(&otherEntsBody); err != nil {
		t.Fatalf("decode other entitlements: %v", err)
	}
	if len(otherEntsBody.Entitlements) != 0 {
		t.Fatalf("other user entitlements = %+v, want none", otherEntsBody.Entitlements)
	}
}

func TestOrderCheckoutWithoutPlanSnapshotsAppPrice(t *testing.T) {
	db, app := newAuthTestApp(t)
	devID := insertVerifiedUser(t, db, "dev@example.com", "old-password")
	insertVerifiedUser(t, db, "buyer@example.com", "old-password")
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

	body := `{"name":"Mini Tool","slug":"mini-tool","tagline":"Small utility.","description":"A small developer utility.","category":"templates","priceCents":1200,"currency":"USD","iconUrl":"","coverImageUrl":"","demoUrl":"","docsUrl":"","sourceUrl":"https://github.com/example/mini-tool","supportUrl":"","tags":[],"version":"0.1.0","releaseNotes":"First release."}`
	submitted := performJSONRequestWithToken(t, app, http.MethodPost, "/api/v1/apps", body, developerToken)
	if submitted.Code != http.StatusCreated {
		t.Fatalf("publish status = %d, want %d; body = %s", submitted.Code, http.StatusCreated, submitted.Body.String())
	}
	var submittedBody struct {
		App struct {
			ID   int64  `json:"id"`
			Slug string `json:"slug"`
		} `json:"app"`
	}
	if err := json.NewDecoder(submitted.Body).Decode(&submittedBody); err != nil {
		t.Fatalf("decode published app: %v", err)
	}
	approved := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/admin/apps/"+strconv.FormatInt(submittedBody.App.ID, 10)+"/approve",
		`{"reviewNote":"OK."}`,
		adminToken,
	)
	if approved.Code != http.StatusOK {
		t.Fatalf("approve app status = %d, want %d", approved.Code, http.StatusOK)
	}

	checkout := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/marketplace/apps/mini-tool/checkout",
		`{}`,
		buyerToken,
	)
	if checkout.Code != http.StatusCreated {
		t.Fatalf("checkout status = %d, want %d; body = %s", checkout.Code, http.StatusCreated, checkout.Body.String())
	}
	var checkoutBody struct {
		Order struct {
			Status     string `json:"status"`
			PlanName   string `json:"planName"`
			PriceCents int64  `json:"priceCents"`
			Currency   string `json:"currency"`
		} `json:"order"`
	}
	if err := json.NewDecoder(checkout.Body).Decode(&checkoutBody); err != nil {
		t.Fatalf("decode checkout: %v", err)
	}
	if checkoutBody.Order.Status != "pending" || checkoutBody.Order.PlanName != "" ||
		checkoutBody.Order.PriceCents != 1200 || checkoutBody.Order.Currency != "USD" {
		t.Fatalf("checkout order = %+v, want pending app price 1200 USD", checkoutBody.Order)
	}
}

func TestOrderCheckoutRejectsForeignPlan(t *testing.T) {
	db, app := newAuthTestApp(t)
	devID := insertVerifiedUser(t, db, "dev@example.com", "old-password")
	insertVerifiedUser(t, db, "buyer@example.com", "old-password")
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

	// Two apps; the second gets a plan.
	publish := func(name, slug string) int64 {
		requestBody := `{"name":"` + name + `","slug":"` + slug + `","tagline":"T.","description":"D.","category":"plugins","priceCents":0,"currency":"USD","iconUrl":"","coverImageUrl":"","demoUrl":"","docsUrl":"","sourceUrl":"https://github.com/example/x","supportUrl":"","tags":[],"version":"1.0.0","releaseNotes":"R.","plans":[{"name":"Plus","priceCents":999,"currency":"USD","description":"P","features":[]}]}`
		resp := performJSONRequestWithToken(t, app, http.MethodPost, "/api/v1/apps", requestBody, developerToken)
		if resp.Code != http.StatusCreated {
			t.Fatalf("publish %s status = %d, want %d", slug, resp.Code, http.StatusCreated)
		}
		var publishBody struct {
			App struct {
				ID int64 `json:"id"`
			} `json:"app"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&publishBody); err != nil {
			t.Fatalf("decode publish %s: %v", slug, err)
		}
		approved := performJSONRequestWithToken(
			t,
			app,
			http.MethodPost,
			"/api/v1/admin/apps/"+strconv.FormatInt(publishBody.App.ID, 10)+"/approve",
			`{"reviewNote":"OK."}`,
			adminToken,
		)
		if approved.Code != http.StatusOK {
			t.Fatalf("approve %s status = %d, want %d", slug, approved.Code, http.StatusOK)
		}
		return publishBody.App.ID
	}
	publish("App One", "app-one")
	publish("App Two", "app-two")

	// Grab the plan ID of app-two.
	detail := performJSONRequest(t, app, http.MethodGet, "/api/v1/marketplace/apps/app-two", "")
	var detailBody struct {
		App struct {
			Plans []struct {
				ID int64 `json:"id"`
			} `json:"plans"`
		} `json:"app"`
	}
	if err := json.NewDecoder(detail.Body).Decode(&detailBody); err != nil {
		t.Fatalf("decode detail: %v", err)
	}
	if len(detailBody.App.Plans) != 1 {
		t.Fatalf("app-two plans = %+v, want 1", detailBody.App.Plans)
	}
	foreignPlanID := detailBody.App.Plans[0].ID

	// Checking out app-one with app-two's plan must be rejected.
	checkout := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/marketplace/apps/app-one/checkout",
		fmt.Sprintf(`{"planId":%d}`, foreignPlanID),
		buyerToken,
	)
	if checkout.Code != http.StatusBadRequest {
		t.Fatalf("foreign plan checkout status = %d, want %d; body = %s", checkout.Code, http.StatusBadRequest, checkout.Body.String())
	}
}
