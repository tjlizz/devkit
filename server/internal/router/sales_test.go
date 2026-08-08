package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
)

func TestDeveloperSalesScopedToOwnApps(t *testing.T) {
	db, app := newAuthTestApp(t)
	devID := insertVerifiedUser(t, db, "dev@example.com", "old-password")
	buyerID := insertVerifiedUser(t, db, "buyer@example.com", "old-password")
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

	// Publish and approve two apps, then buy one of them.
	appOneID := publishAndApprove(t, app, developerToken, adminToken, "Sales App One", "sales-app-one")
	_ = publishAndApprove(t, app, developerToken, adminToken, "Sales App Two", "sales-app-two")

	checkout := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/marketplace/apps/sales-app-one/checkout",
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
	confirm := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/orders/"+strconv.FormatInt(checkoutBody.Order.ID, 10)+"/confirm-payment",
		"",
		buyerToken,
	)
	if confirm.Code != http.StatusOK {
		t.Fatalf("confirm status = %d, want %d; body = %s", confirm.Code, http.StatusOK, confirm.Body.String())
	}

	// --- Unauthenticated request is rejected. ---
	unauth := performJSONRequest(t, app, http.MethodGet, "/api/v1/developer/sales", "")
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated sales status = %d, want %d", unauth.Code, http.StatusUnauthorized)
	}

	// --- A plain buyer (not a developer) is forbidden. ---
	buyerForbidden := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/developer/sales", "", buyerToken)
	if buyerForbidden.Code != http.StatusForbidden {
		t.Fatalf("buyer sales status = %d, want %d", buyerForbidden.Code, http.StatusForbidden)
	}

	// --- The developer sees only their own paid order across their apps. ---
	sales := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/developer/sales", "", developerToken)
	if sales.Code != http.StatusOK {
		t.Fatalf("sales status = %d, want %d; body = %s", sales.Code, http.StatusOK, sales.Body.String())
	}
	var salesBody struct {
		Sales []struct {
			OrderID    int64  `json:"orderId"`
			AppID      int64  `json:"appId"`
			AppSlug    string `json:"appSlug"`
			AppName    string `json:"appName"`
			BuyerEmail string `json:"buyerEmail"`
			PriceCents int64  `json:"priceCents"`
			Currency   string `json:"currency"`
			PaidAt     string `json:"paidAt"`
		} `json:"sales"`
		Summary struct {
			TotalOrders       int64 `json:"totalOrders"`
			TotalRevenueCents int64 `json:"totalRevenueCents"`
			UniqueBuyers      int64 `json:"uniqueBuyers"`
			AppsSold          int64 `json:"appsSold"`
		} `json:"summary"`
	}
	if err := json.NewDecoder(sales.Body).Decode(&salesBody); err != nil {
		t.Fatalf("decode sales: %v", err)
	}
	if len(salesBody.Sales) != 1 {
		t.Fatalf("sales len = %d, want 1; body = %s", len(salesBody.Sales), sales.Body.String())
	}
	sale := salesBody.Sales[0]
	if sale.AppID != appOneID || sale.AppSlug != "sales-app-one" || sale.AppName != "Sales App One" {
		t.Fatalf("sale app = %+v, want app-one", sale)
	}
	if sale.BuyerEmail != "buyer@example.com" {
		t.Fatalf("sale buyer email = %q, want buyer@example.com", sale.BuyerEmail)
	}
	if sale.PriceCents != 2900 || sale.Currency != "USD" || sale.PaidAt == "" {
		t.Fatalf("sale price/currency/paidAt = %+v, want 2900 USD paid", sale)
	}
	if salesBody.Summary.TotalOrders != 1 || salesBody.Summary.TotalRevenueCents != 2900 ||
		salesBody.Summary.UniqueBuyers != 1 || salesBody.Summary.AppsSold != 1 {
		t.Fatalf("summary = %+v, want 1 order / 2900 / 1 buyer / 1 app", salesBody.Summary)
	}

	// A second developer with sales must not see this developer's orders.
	otherDevID := insertVerifiedUser(t, db, "other-dev@example.com", "old-password")
	if _, err := db.Exec("INSERT INTO developers(user_id, display_name) VALUES (?, ?)", otherDevID, "Other Studio"); err != nil {
		t.Fatalf("insert other developer: %v", err)
	}
	otherToken := loginToken(t, app, "other-dev@example.com", "old-password")
	otherSales := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/developer/sales", "", otherToken)
	if otherSales.Code != http.StatusOK {
		t.Fatalf("other dev sales status = %d, want %d", otherSales.Code, http.StatusOK)
	}
	var otherBody struct {
		Sales   []json.RawMessage `json:"sales"`
		Summary struct {
			TotalOrders int64 `json:"totalOrders"`
		} `json:"summary"`
	}
	if err := json.NewDecoder(otherSales.Body).Decode(&otherBody); err != nil {
		t.Fatalf("decode other dev sales: %v", err)
	}
	if len(otherBody.Sales) != 0 || otherBody.Summary.TotalOrders != 0 {
		t.Fatalf("other dev sales = %+v, want empty", otherBody)
	}

	// Sanity: buyer id is genuinely the buyer's user id.
	if buyerID < 1 {
		t.Fatalf("buyer id = %d, want positive", buyerID)
	}
	_ = fmt.Sprintf("%d", appOneID) // keep appOneID referenced
}
