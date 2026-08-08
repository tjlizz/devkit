package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"devkit/server/internal/middleware"

	"github.com/google/uuid"
)

// deliveryTokenLifetime bounds how long a signed delivery token stays valid.
const deliveryTokenLifetime = 15 * time.Minute

type checkoutRequest struct {
	PlanID *int64 `json:"planId"`
}

type order struct {
	ID              int64  `json:"id"`
	BuyerID         int64  `json:"buyerId"`
	AppID           int64  `json:"appId"`
	AppSlug         string `json:"appSlug"`
	AppName         string `json:"appName"`
	PlanID          *int64 `json:"planId,omitempty"`
	PlanName        string `json:"planName"`
	PriceCents      int64  `json:"priceCents"`
	Currency        string `json:"currency"`
	Status          string `json:"status"`
	Provider        string `json:"provider"`
	ProviderEventID string `json:"providerEventId"`
	PaidAt          string `json:"paidAt,omitempty"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

type entitlement struct {
	ID        int64  `json:"id"`
	BuyerID   int64  `json:"buyerId"`
	AppID     int64  `json:"appId"`
	AppSlug   string `json:"appSlug"`
	AppName   string `json:"appName"`
	PlanID    *int64 `json:"planId,omitempty"`
	PlanName  string `json:"planName"`
	OrderID   int64  `json:"orderId"`
	Status    string `json:"status"`
	GrantedAt string `json:"grantedAt"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// Checkout creates a pending order for an approved marketplace app. The price
// is snapshotted from the selected plan (or the app itself) at order time so a
// later price change cannot alter a placed order.
func (h *Auth) Checkout(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	slug := strings.TrimSpace(r.PathValue("slug"))
	if slug == "" {
		writeError(w, http.StatusBadRequest, "invalid app slug")
		return
	}
	app, err := h.readAppBySlug(r, slug)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read app for checkout", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read app")
		return
	}

	var request checkoutRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var planID *int64
	var planName string
	priceCents := app.PriceCents
	currency := app.Currency

	if request.PlanID != nil {
		plan, err := h.readPlan(r, *request.PlanID)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "plan not found")
			return
		}
		if err != nil {
			h.logger.ErrorContext(r.Context(), "read plan for checkout", "error", err)
			writeError(w, http.StatusInternalServerError, "could not read plan")
			return
		}
		if plan.AppID != app.ID {
			writeError(w, http.StatusBadRequest, "plan does not belong to this app")
			return
		}
		id := plan.ID
		planID = &id
		planName = plan.Name
		priceCents = plan.PriceCents
		currency = plan.Currency
	}

	// One active entitlement per buyer per app: reject duplicate purchases.
	var existing int64
	if err := h.db.QueryRowContext(
		r.Context(),
		`SELECT COUNT(*) FROM entitlements WHERE buyer_id = ? AND app_id = ? AND status = 'active'`,
		userID, app.ID,
	).Scan(&existing); err != nil {
		h.logger.ErrorContext(r.Context(), "check existing entitlement", "error", err)
		writeError(w, http.StatusInternalServerError, "could not verify entitlement")
		return
	}
	if existing > 0 {
		writeError(w, http.StatusConflict, "you already own this app")
		return
	}

	var orderID int64
	if err := h.db.QueryRowContext(
		r.Context(),
		`INSERT INTO orders(buyer_id, app_id, plan_id, plan_name, price_cents, currency, status, provider)
		 VALUES (?, ?, ?, ?, ?, ?, 'pending', 'sandbox')
		 RETURNING id`,
		userID, app.ID, planID, planName, priceCents, currency,
	).Scan(&orderID); err != nil {
		h.logger.ErrorContext(r.Context(), "insert order", "error", err)
		writeError(w, http.StatusInternalServerError, "could not create order")
		return
	}

	created, err := h.readOrder(r, orderID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read created order", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read order")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]order{"order": created})
}

// ConfirmPayment transitions a pending order to paid and issues the
// entitlement from that confirmed server-side payment state. It is idempotent:
// confirming an already-paid order returns the same result without re-issuing.
func (h *Auth) ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	orderID, ok := orderIDFromPath(w, r)
	if !ok {
		return
	}

	current, err := h.readOrder(r, orderID)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read order for payment", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read order")
		return
	}
	if current.BuyerID != userID {
		writeError(w, http.StatusForbidden, "order does not belong to you")
		return
	}

	// Idempotent: a paid order returns the existing entitlement as-is.
	if current.Status == "paid" {
		ent, err := h.readEntitlementByOrder(r, orderID)
		if err != nil {
			h.logger.ErrorContext(r.Context(), "read paid entitlement", "error", err)
			writeError(w, http.StatusInternalServerError, "could not read entitlement")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"order": current, "entitlement": ent})
		return
	}
	if current.Status != "pending" {
		writeError(w, http.StatusConflict, "order is already "+current.Status)
		return
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "begin payment transaction", "error", err)
		writeError(w, http.StatusInternalServerError, "could not confirm payment")
		return
	}
	defer tx.Rollback()

	eventID := uuid.NewString()
	now := time.Now().UTC()
	result, err := tx.ExecContext(
		r.Context(),
		`UPDATE orders
		 SET status = 'paid', provider_event_id = ?, paid_at = ?, updated_at = ?
		 WHERE id = ? AND status = 'pending'`,
		eventID, now, now, orderID,
	)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "update order to paid", "error", err)
		writeError(w, http.StatusInternalServerError, "could not confirm payment")
		return
	}
	affected, err := result.RowsAffected()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read update result", "error", err)
		writeError(w, http.StatusInternalServerError, "could not confirm payment")
		return
	}
	if affected != 1 {
		writeError(w, http.StatusConflict, "order is already "+current.Status)
		return
	}

	if _, err := tx.ExecContext(
		r.Context(),
		`INSERT INTO entitlements(buyer_id, app_id, plan_id, order_id, status)
		 VALUES (?, ?, ?, ?, 'active')`,
		current.BuyerID, current.AppID, current.PlanID, orderID,
	); err != nil {
		h.logger.ErrorContext(r.Context(), "insert entitlement", "error", err)
		writeError(w, http.StatusInternalServerError, "could not issue entitlement")
		return
	}

	if err := tx.Commit(); err != nil {
		h.logger.ErrorContext(r.Context(), "commit payment transaction", "error", err)
		writeError(w, http.StatusInternalServerError, "could not confirm payment")
		return
	}

	updated, err := h.readOrder(r, orderID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read paid order", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read order")
		return
	}
	ent, err := h.readEntitlementByOrder(r, orderID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read issued entitlement", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read entitlement")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"order": updated, "entitlement": ent})
}

// ListMyOrders returns the authenticated buyer's orders, newest first.
func (h *Auth) ListMyOrders(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	orders, err := h.queryOrders(r, `WHERE o.buyer_id = ? ORDER BY o.id DESC`, userID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "list my orders", "error", err)
		writeError(w, http.StatusInternalServerError, "could not list orders")
		return
	}
	writeJSON(w, http.StatusOK, map[string][]order{"orders": orders})
}

// developerSale is a single paid order belonging to one of the developer's apps.
type developerSale struct {
	OrderID    int64  `json:"orderId"`
	AppID      int64  `json:"appId"`
	AppSlug    string `json:"appSlug"`
	AppName    string `json:"appName"`
	PlanName   string `json:"planName"`
	BuyerEmail string `json:"buyerEmail"`
	PriceCents int64  `json:"priceCents"`
	Currency   string `json:"currency"`
	Provider   string `json:"provider"`
	PaidAt     string `json:"paidAt,omitempty"`
	CreatedAt  string `json:"createdAt"`
}

// developerSalesSummary aggregates the developer's paid orders.
type developerSalesSummary struct {
	TotalOrders       int64 `json:"totalOrders"`
	TotalRevenueCents int64 `json:"totalRevenueCents"`
	UniqueBuyers      int64 `json:"uniqueBuyers"`
	AppsSold          int64 `json:"appsSold"`
}

// ListDeveloperSales returns the authenticated developer's paid orders across
// all of their apps, newest first, plus a summary of paid orders, revenue,
// unique buyers, and apps sold.
func (h *Auth) ListDeveloperSales(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	developerID, ok := h.requireDeveloperID(w, r, userID)
	if !ok {
		return
	}

	limit, offset := pagination(r)
	rows, err := h.db.QueryContext(
		r.Context(),
		`SELECT o.id, o.app_id, a.slug, a.name, COALESCE(o.plan_name, ''), u.email,
		        o.price_cents, o.currency, o.provider, o.paid_at, o.created_at
		 FROM orders o
		 JOIN apps a ON a.id = o.app_id
		 JOIN users u ON u.id = o.buyer_id
		 WHERE a.developer_id = ? AND o.status = 'paid'
		 ORDER BY o.id DESC LIMIT ? OFFSET ?`,
		developerID, limit, offset,
	)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "list developer sales", "error", err)
		writeError(w, http.StatusInternalServerError, "could not list sales")
		return
	}
	defer rows.Close()

	sales := make([]developerSale, 0)
	for rows.Next() {
		var sale developerSale
		var paidAt sql.NullTime
		var createdAt time.Time
		if err := rows.Scan(
			&sale.OrderID, &sale.AppID, &sale.AppSlug, &sale.AppName, &sale.PlanName,
			&sale.BuyerEmail, &sale.PriceCents, &sale.Currency, &sale.Provider,
			&paidAt, &createdAt,
		); err != nil {
			h.logger.ErrorContext(r.Context(), "scan developer sale", "error", err)
			writeError(w, http.StatusInternalServerError, "could not list sales")
			return
		}
		if paidAt.Valid {
			sale.PaidAt = paidAt.Time.UTC().Format(time.RFC3339)
		}
		sale.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		sales = append(sales, sale)
	}
	if err := rows.Err(); err != nil {
		h.logger.ErrorContext(r.Context(), "iterate developer sales", "error", err)
		writeError(w, http.StatusInternalServerError, "could not list sales")
		return
	}

	var summary developerSalesSummary
	if err := h.db.QueryRowContext(
		r.Context(),
		`SELECT COUNT(*), COALESCE(SUM(o.price_cents), 0), COUNT(DISTINCT o.buyer_id), COUNT(DISTINCT o.app_id)
		 FROM orders o
		 JOIN apps a ON a.id = o.app_id
		 WHERE a.developer_id = ? AND o.status = 'paid'`,
		developerID,
	).Scan(&summary.TotalOrders, &summary.TotalRevenueCents, &summary.UniqueBuyers, &summary.AppsSold); err != nil {
		h.logger.ErrorContext(r.Context(), "summarize developer sales", "error", err)
		writeError(w, http.StatusInternalServerError, "could not summarize sales")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"sales":   sales,
		"summary": summary,
		"limit":   limit,
		"offset":  offset,
	})
}

// ListMyEntitlements returns the authenticated buyer's entitlements.
func (h *Auth) ListMyEntitlements(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	ents, err := h.queryEntitlements(r, `WHERE e.buyer_id = ? ORDER BY e.id DESC`, userID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "list my entitlements", "error", err)
		writeError(w, http.StatusInternalServerError, "could not list entitlements")
		return
	}
	writeJSON(w, http.StatusOK, map[string][]entitlement{"entitlements": ents})
}

// GetDelivery returns a short-lived signed delivery token for an owned app.
// The token is meant for artifact access without exposing the entitlement
// itself; it expires after deliveryTokenLifetime.
func (h *Auth) GetDelivery(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	entID, ok := entitlementIDFromPath(w, r)
	if !ok {
		return
	}
	ent, err := h.readEntitlement(r, entID)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "entitlement not found")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read entitlement for delivery", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read entitlement")
		return
	}
	if ent.BuyerID != userID {
		writeError(w, http.StatusForbidden, "entitlement does not belong to you")
		return
	}
	if ent.Status != "active" {
		writeError(w, http.StatusForbidden, "entitlement is not active")
		return
	}

	expiresAt := time.Now().UTC().Add(deliveryTokenLifetime)
	token, err := h.signDeliveryToken(r, ent.ID, expiresAt)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "sign delivery token", "error", err)
		writeError(w, http.StatusInternalServerError, "could not create delivery token")
		return
	}

	app, err := h.readApp(r, ent.AppID, false)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read app for delivery", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read app")
		return
	}

	artifacts, err := h.listArtifactsRaw(r, app.ID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "list artifacts for delivery", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read artifacts")
		return
	}

	publicArtifacts := make([]appArtifact, 0, len(artifacts))
	for _, meta := range artifacts {
		publicArtifacts = append(publicArtifacts, toAppArtifact(meta))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"entitlementId": ent.ID,
		"appSlug":       app.Slug,
		"appName":       app.Name,
		"version":       app.Version,
		"sourceUrl":     app.SourceURL,
		"docsUrl":       app.DocsURL,
		"demoUrl":       app.DemoURL,
		"artifacts":     publicArtifacts,
		"deliveryToken": token,
		"expiresAt":     expiresAt.Format(time.RFC3339),
	})
}

func (h *Auth) readPlan(r *http.Request, planID int64) (appPlan, error) {
	var plan appPlan
	var featuresJSON string
	var createdAt, updatedAt time.Time
	err := h.db.QueryRowContext(
		r.Context(),
		`SELECT id, app_id, name, price_cents, currency, description, features, sort_order,
		        created_at, updated_at
		 FROM app_plans WHERE id = ?`,
		planID,
	).Scan(
		&plan.ID, &plan.AppID, &plan.Name, &plan.PriceCents, &plan.Currency, &plan.Description,
		&featuresJSON, &plan.SortOrder, &createdAt, &updatedAt,
	)
	if err != nil {
		return appPlan{}, err
	}
	if err := json.Unmarshal([]byte(featuresJSON), &plan.Features); err != nil {
		plan.Features = []string{}
	}
	plan.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	plan.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
	return plan, nil
}

func (h *Auth) readOrder(r *http.Request, orderID int64) (order, error) {
	return h.scanOrder(h.db.QueryRowContext(
		r.Context(),
		`SELECT o.id, o.buyer_id, o.app_id, a.slug, a.name, o.plan_id, o.plan_name,
		        o.price_cents, o.currency, o.status, o.provider, o.provider_event_id,
		        o.paid_at, o.created_at, o.updated_at
		 FROM orders o
		 JOIN apps a ON a.id = o.app_id
		 WHERE o.id = ?`,
		orderID,
	))
}

func (h *Auth) queryOrders(r *http.Request, where string, args ...any) ([]order, error) {
	rows, err := h.db.QueryContext(
		r.Context(),
		`SELECT o.id, o.buyer_id, o.app_id, a.slug, a.name, o.plan_id, o.plan_name,
		        o.price_cents, o.currency, o.status, o.provider, o.provider_event_id,
		        o.paid_at, o.created_at, o.updated_at
		 FROM orders o
		 JOIN apps a ON a.id = o.app_id `+where,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orders := make([]order, 0)
	for rows.Next() {
		item, err := h.scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, item)
	}
	return orders, rows.Err()
}

func (h *Auth) scanOrder(scanner appScanner) (order, error) {
	var item order
	var planID sql.NullInt64
	var paidAt sql.NullTime
	var createdAt, updatedAt time.Time
	err := scanner.Scan(
		&item.ID, &item.BuyerID, &item.AppID, &item.AppSlug, &item.AppName,
		&planID, &item.PlanName, &item.PriceCents, &item.Currency, &item.Status,
		&item.Provider, &item.ProviderEventID, &paidAt, &createdAt, &updatedAt,
	)
	if err != nil {
		return order{}, err
	}
	if planID.Valid {
		value := planID.Int64
		item.PlanID = &value
	}
	if paidAt.Valid {
		item.PaidAt = paidAt.Time.UTC().Format(time.RFC3339)
	}
	item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
	return item, nil
}

func (h *Auth) readEntitlementByOrder(r *http.Request, orderID int64) (entitlement, error) {
	return h.scanEntitlement(h.db.QueryRowContext(
		r.Context(),
		`SELECT e.id, e.buyer_id, e.app_id, a.slug, a.name, e.plan_id, COALESCE(o.plan_name, ''),
		        e.order_id, e.status, e.granted_at, e.created_at, e.updated_at
		 FROM entitlements e
		 JOIN apps a ON a.id = e.app_id
		 JOIN orders o ON o.id = e.order_id
		 WHERE e.order_id = ?`,
		orderID,
	))
}

func (h *Auth) readEntitlement(r *http.Request, entID int64) (entitlement, error) {
	return h.scanEntitlement(h.db.QueryRowContext(
		r.Context(),
		`SELECT e.id, e.buyer_id, e.app_id, a.slug, a.name, e.plan_id, COALESCE(o.plan_name, ''),
		        e.order_id, e.status, e.granted_at, e.created_at, e.updated_at
		 FROM entitlements e
		 JOIN apps a ON a.id = e.app_id
		 JOIN orders o ON o.id = e.order_id
		 WHERE e.id = ?`,
		entID,
	))
}

func (h *Auth) queryEntitlements(r *http.Request, where string, args ...any) ([]entitlement, error) {
	rows, err := h.db.QueryContext(
		r.Context(),
		`SELECT e.id, e.buyer_id, e.app_id, a.slug, a.name, e.plan_id, COALESCE(o.plan_name, ''),
		        e.order_id, e.status, e.granted_at, e.created_at, e.updated_at
		 FROM entitlements e
		 JOIN apps a ON a.id = e.app_id
		 JOIN orders o ON o.id = e.order_id `+where,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ents := make([]entitlement, 0)
	for rows.Next() {
		item, err := h.scanEntitlement(rows)
		if err != nil {
			return nil, err
		}
		ents = append(ents, item)
	}
	return ents, rows.Err()
}

func (h *Auth) scanEntitlement(scanner appScanner) (entitlement, error) {
	var item entitlement
	var planID sql.NullInt64
	var grantedAt, createdAt, updatedAt time.Time
	err := scanner.Scan(
		&item.ID, &item.BuyerID, &item.AppID, &item.AppSlug, &item.AppName,
		&planID, &item.PlanName, &item.OrderID, &item.Status,
		&grantedAt, &createdAt, &updatedAt,
	)
	if err != nil {
		return entitlement{}, err
	}
	if planID.Valid {
		value := planID.Int64
		item.PlanID = &value
	}
	item.GrantedAt = grantedAt.UTC().Format(time.RFC3339)
	item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
	return item, nil
}

// signDeliveryToken returns a URL-safe HMAC token binding the entitlement ID
// to an expiry, so delivery endpoints can be called without the JWT.
func (h *Auth) signDeliveryToken(r *http.Request, entitlementID int64, expiresAt time.Time) (string, error) {
	payload := fmt.Sprintf("%d:%d", entitlementID, expiresAt.Unix())
	mac := hmac.New(sha256.New, []byte(h.config.JWTSecret))
	if _, err := mac.Write([]byte(payload)); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(append([]byte(payload+":"), mac.Sum(nil)...)), nil
}

func orderIDFromPath(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return 0, false
	}
	return id, true
}

func entitlementIDFromPath(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid entitlement id")
		return 0, false
	}
	return id, true
}
