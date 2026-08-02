package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"devkit/server/internal/middleware"
)

var appCategories = map[string]bool{
	"saas":            true,
	"ai-applications": true,
	"developer-tools": true,
	"templates":       true,
	"plugins":         true,
	"apis":            true,
	"open-source":     true,
}

type appPlanInput struct {
	Name        string   `json:"name"`
	PriceCents  int64    `json:"priceCents"`
	Currency    string   `json:"currency"`
	Description string   `json:"description"`
	Features    []string `json:"features"`
}

type appPlan struct {
	ID          int64    `json:"id"`
	AppID       int64    `json:"appId"`
	Name        string   `json:"name"`
	PriceCents  int64    `json:"priceCents"`
	Currency    string   `json:"currency"`
	Description string   `json:"description"`
	Features    []string `json:"features"`
	SortOrder   int      `json:"sortOrder"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

type appPublishRequest struct {
	Name          string         `json:"name"`
	Slug          string         `json:"slug"`
	Tagline       string         `json:"tagline"`
	Description   string         `json:"description"`
	Category      string         `json:"category"`
	PriceCents    int64          `json:"priceCents"`
	Currency      string         `json:"currency"`
	IconURL       string         `json:"iconUrl"`
	CoverImageURL string         `json:"coverImageUrl"`
	DemoURL       string         `json:"demoUrl"`
	DocsURL       string         `json:"docsUrl"`
	SourceURL     string         `json:"sourceUrl"`
	SupportURL    string         `json:"supportUrl"`
	Tags          []string       `json:"tags"`
	Version       string         `json:"version"`
	ReleaseNotes  string         `json:"releaseNotes"`
	Plans         []appPlanInput `json:"plans"`
}

type appReviewRequest struct {
	ReviewNote string `json:"reviewNote"`
}

type marketplaceApp struct {
	ID             int64     `json:"id"`
	DeveloperID    int64     `json:"developerId"`
	DeveloperName  string    `json:"developerName"`
	DeveloperEmail string    `json:"developerEmail,omitempty"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Tagline        string    `json:"tagline"`
	Description    string    `json:"description"`
	Category       string    `json:"category"`
	PriceCents     int64     `json:"priceCents"`
	Currency       string    `json:"currency"`
	IconURL        string    `json:"iconUrl"`
	CoverImageURL  string    `json:"coverImageUrl"`
	DemoURL        string    `json:"demoUrl"`
	DocsURL        string    `json:"docsUrl"`
	SourceURL      string    `json:"sourceUrl"`
	SupportURL     string    `json:"supportUrl"`
	Tags           []string  `json:"tags"`
	Version        string    `json:"version"`
	ReleaseNotes   string    `json:"releaseNotes"`
	Status         string    `json:"status"`
	ReviewNote     string    `json:"reviewNote"`
	ReviewedBy     *int64    `json:"reviewedBy,omitempty"`
	ReviewedAt     string    `json:"reviewedAt,omitempty"`
	PublishedAt    string    `json:"publishedAt,omitempty"`
	CreatedAt      string    `json:"createdAt"`
	UpdatedAt      string    `json:"updatedAt"`
	FavoriteCount  int       `json:"favoriteCount"`
	Plans          []appPlan `json:"plans"`
}

type marketplaceFavorite struct {
	Favorited     bool `json:"favorited"`
	FavoriteCount int  `json:"favoriteCount"`
}

func (h *Auth) CreateApp(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	developerID, ok := h.requireDeveloperID(w, r, userID)
	if !ok {
		return
	}

	var request appPublishRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.createAppForDeveloper(w, r, developerID, request)
}

func (h *Auth) createAppForDeveloper(w http.ResponseWriter, r *http.Request, developerID int64, request appPublishRequest) {
	normalizeAppPublishRequest(&request)
	if err := validateAppPublishRequest(request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	tagsJSON, err := json.Marshal(request.Tags)
	if err != nil {
		writeError(w, http.StatusBadRequest, "tags are invalid")
		return
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "begin create app tx", "error", err)
		writeError(w, http.StatusInternalServerError, "could not publish app")
		return
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(
		r.Context(),
		`INSERT INTO apps(developer_id, name, slug, tagline, description, category, price_cents,
		    currency, icon_url, cover_image_url, demo_url, docs_url, source_url, support_url,
		    tags, version, release_notes, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending_review')`,
		developerID, request.Name, request.Slug, request.Tagline, request.Description, request.Category,
		request.PriceCents, request.Currency, request.IconURL, request.CoverImageURL, request.DemoURL,
		request.DocsURL, request.SourceURL, request.SupportURL, string(tagsJSON), request.Version,
		request.ReleaseNotes,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: apps.slug") {
			writeError(w, http.StatusConflict, "app slug is already taken")
			return
		}
		h.logger.ErrorContext(r.Context(), "create app", "error", err)
		writeError(w, http.StatusInternalServerError, "could not publish app")
		return
	}
	appID, err := result.LastInsertId()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read app id", "error", err)
		writeError(w, http.StatusInternalServerError, "could not publish app")
		return
	}
	if err := h.replacePlans(r, tx, appID, request.Plans); err != nil {
		h.logger.ErrorContext(r.Context(), "insert app plans", "error", err)
		writeError(w, http.StatusInternalServerError, "could not publish app")
		return
	}
	if err := tx.Commit(); err != nil {
		h.logger.ErrorContext(r.Context(), "commit create app tx", "error", err)
		writeError(w, http.StatusInternalServerError, "could not publish app")
		return
	}
	app, err := h.readApp(r, appID, false)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read created app", "error", err)
		writeError(w, http.StatusInternalServerError, "could not publish app")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]marketplaceApp{"app": app})
}

func (h *Auth) ListDeveloperApps(w http.ResponseWriter, r *http.Request) {
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
	apps, err := h.queryApps(
		r,
		`SELECT a.id, a.developer_id, d.display_name, '', a.name, a.slug, a.tagline,
		        a.description, a.category, a.price_cents, a.currency, a.icon_url, a.cover_image_url,
		        a.demo_url, a.docs_url, a.source_url, a.support_url, a.tags, a.version,
		        a.release_notes, a.status, a.review_note, a.reviewed_by, a.reviewed_at,
		        a.published_at, a.created_at, a.updated_at
		 FROM apps a
		 JOIN developers d ON d.id = a.developer_id
		 WHERE a.developer_id = ?
		 ORDER BY a.updated_at DESC, a.id DESC LIMIT ? OFFSET ?`,
		false,
		developerID, limit, offset,
	)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "list developer apps", "error", err)
		writeError(w, http.StatusInternalServerError, "could not list apps")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"apps": apps, "limit": limit, "offset": offset})
}

func (h *Auth) GetDeveloperApp(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	developerID, ok := h.requireDeveloperID(w, r, userID)
	if !ok {
		return
	}
	appID, ok := appIDFromPath(w, r)
	if !ok {
		return
	}
	app, err := h.readDeveloperApp(r, appID, developerID)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read developer app", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read app")
		return
	}
	writeJSON(w, http.StatusOK, map[string]marketplaceApp{"app": app})
}

func (h *Auth) UpdateDeveloperApp(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	developerID, ok := h.requireDeveloperID(w, r, userID)
	if !ok {
		return
	}
	appID, ok := appIDFromPath(w, r)
	if !ok {
		return
	}
	current, err := h.readDeveloperApp(r, appID, developerID)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read app before update", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update app")
		return
	}
	if current.Status == "delisted" {
		writeError(w, http.StatusConflict, "delisted apps cannot be edited")
		return
	}

	var request appPublishRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	normalizeAppPublishRequest(&request)
	if err := validateAppPublishRequest(request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	tagsJSON, err := json.Marshal(request.Tags)
	if err != nil {
		writeError(w, http.StatusBadRequest, "tags are invalid")
		return
	}

	nextStatus := current.Status
	if current.Status == "approved" {
		nextStatus = "pending_review"
	}
	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "begin update app tx", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update app")
		return
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(
		r.Context(),
		`UPDATE apps
		 SET name = ?, slug = ?, tagline = ?, description = ?, category = ?, price_cents = ?,
		     currency = ?, icon_url = ?, cover_image_url = ?, demo_url = ?, docs_url = ?,
		     source_url = ?, support_url = ?, tags = ?, version = ?, release_notes = ?,
		     status = ?, review_note = CASE WHEN ? = 'pending_review' THEN '' ELSE review_note END,
		     reviewed_by = CASE WHEN ? = 'pending_review' THEN NULL ELSE reviewed_by END,
		     reviewed_at = CASE WHEN ? = 'pending_review' THEN NULL ELSE reviewed_at END,
		     published_at = CASE WHEN ? = 'pending_review' THEN NULL ELSE published_at END,
		     updated_at = CURRENT_TIMESTAMP
		 WHERE id = ? AND developer_id = ?`,
		request.Name, request.Slug, request.Tagline, request.Description, request.Category, request.PriceCents,
		request.Currency, request.IconURL, request.CoverImageURL, request.DemoURL, request.DocsURL,
		request.SourceURL, request.SupportURL, string(tagsJSON), request.Version, request.ReleaseNotes,
		nextStatus, nextStatus, nextStatus, nextStatus, nextStatus, appID, developerID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: apps.slug") {
			writeError(w, http.StatusConflict, "app slug is already taken")
			return
		}
		h.logger.ErrorContext(r.Context(), "update developer app", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update app")
		return
	}
	if err := h.replacePlans(r, tx, appID, request.Plans); err != nil {
		h.logger.ErrorContext(r.Context(), "replace app plans", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update app")
		return
	}
	if err := tx.Commit(); err != nil {
		h.logger.ErrorContext(r.Context(), "commit update app tx", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update app")
		return
	}
	app, err := h.readDeveloperApp(r, appID, developerID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read updated developer app", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update app")
		return
	}
	writeJSON(w, http.StatusOK, map[string]marketplaceApp{"app": app})
}

func (h *Auth) DelistDeveloperApp(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	developerID, ok := h.requireDeveloperID(w, r, userID)
	if !ok {
		return
	}
	appID, ok := appIDFromPath(w, r)
	if !ok {
		return
	}
	current, err := h.readDeveloperApp(r, appID, developerID)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read app before delist", "error", err)
		writeError(w, http.StatusInternalServerError, "could not delist app")
		return
	}
	if current.Status == "delisted" {
		writeJSON(w, http.StatusOK, map[string]marketplaceApp{"app": current})
		return
	}
	if current.Status != "approved" {
		writeError(w, http.StatusConflict, "only approved apps can be delisted")
		return
	}
	if _, err := h.db.ExecContext(
		r.Context(),
		`UPDATE apps
		 SET status = 'delisted', review_note = '', published_at = NULL, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ? AND developer_id = ?`,
		appID, developerID,
	); err != nil {
		h.logger.ErrorContext(r.Context(), "delist developer app", "error", err)
		writeError(w, http.StatusInternalServerError, "could not delist app")
		return
	}
	app, err := h.readDeveloperApp(r, appID, developerID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read delisted developer app", "error", err)
		writeError(w, http.StatusInternalServerError, "could not delist app")
		return
	}
	writeJSON(w, http.StatusOK, map[string]marketplaceApp{"app": app})
}

func (h *Auth) ListAdminApps(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status == "" {
		status = "pending_review"
	}
	if status != "all" && !validAppStatus(status) {
		writeError(w, http.StatusBadRequest, "invalid app status")
		return
	}
	limit, offset := pagination(r)
	query := `SELECT a.id, a.developer_id, d.display_name, u.email, a.name, a.slug, a.tagline,
	         a.description, a.category, a.price_cents, a.currency, a.icon_url, a.cover_image_url,
	         a.demo_url, a.docs_url, a.source_url, a.support_url, a.tags, a.version,
	         a.release_notes, a.status, a.review_note, a.reviewed_by, a.reviewed_at,
	         a.published_at, a.created_at, a.updated_at
	         FROM apps a
	         JOIN developers d ON d.id = a.developer_id
	         JOIN users u ON u.id = d.user_id`
	args := []any{}
	if status != "all" {
		query += " WHERE a.status = ?"
		args = append(args, status)
	}
	query += " ORDER BY a.created_at DESC, a.id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	apps, err := h.queryApps(r, query, true, args...)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "list admin apps", "error", err)
		writeError(w, http.StatusInternalServerError, "could not list apps")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"apps": apps, "limit": limit, "offset": offset})
}

func (h *Auth) ApproveApp(w http.ResponseWriter, r *http.Request) {
	h.reviewApp(w, r, "approved")
}

func (h *Auth) RejectApp(w http.ResponseWriter, r *http.Request) {
	h.reviewApp(w, r, "rejected")
}

func (h *Auth) reviewApp(w http.ResponseWriter, r *http.Request, nextStatus string) {
	adminID, ok := h.requireAdminID(w, r)
	if !ok {
		return
	}
	appID, ok := appIDFromPath(w, r)
	if !ok {
		return
	}
	var request appReviewRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	request.ReviewNote = strings.TrimSpace(request.ReviewNote)
	if nextStatus == "rejected" && request.ReviewNote == "" {
		writeError(w, http.StatusBadRequest, "reviewNote is required")
		return
	}
	if utf8.RuneCountInString(request.ReviewNote) > 500 {
		writeError(w, http.StatusBadRequest, "reviewNote must not exceed 500 characters")
		return
	}
	setPublished := "published_at"
	if nextStatus == "approved" {
		setPublished = "CURRENT_TIMESTAMP"
	}
	result, err := h.db.ExecContext(
		r.Context(),
		`UPDATE apps
		 SET status = ?, review_note = ?, reviewed_by = ?, reviewed_at = CURRENT_TIMESTAMP,
		     published_at = `+setPublished+`, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ? AND status = 'pending_review'`,
		nextStatus, request.ReviewNote, adminID, appID,
	)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "review app", "error", err)
		writeError(w, http.StatusInternalServerError, "could not review app")
		return
	}
	affected, err := result.RowsAffected()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read reviewed app rows", "error", err)
		writeError(w, http.StatusInternalServerError, "could not review app")
		return
	}
	if affected == 0 {
		app, readErr := h.readApp(r, appID, false)
		if errors.Is(readErr, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "app not found")
			return
		}
		if readErr != nil {
			h.logger.ErrorContext(r.Context(), "read app review conflict", "error", readErr)
			writeError(w, http.StatusInternalServerError, "could not review app")
			return
		}
		writeError(w, http.StatusConflict, "app is already "+app.Status)
		return
	}
	app, err := h.readApp(r, appID, false)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read reviewed app", "error", err)
		writeError(w, http.StatusInternalServerError, "could not review app")
		return
	}
	writeJSON(w, http.StatusOK, map[string]marketplaceApp{"app": app})
}

func (h *Auth) ListMarketplaceApps(w http.ResponseWriter, r *http.Request) {
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	queryText := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	if category != "" && !appCategories[category] {
		writeError(w, http.StatusBadRequest, "invalid category")
		return
	}
	query := `SELECT a.id, a.developer_id, d.display_name, '', a.name, a.slug, a.tagline,
	         a.description, a.category, a.price_cents, a.currency, a.icon_url, a.cover_image_url,
	         a.demo_url, a.docs_url, a.source_url, a.support_url, a.tags, a.version,
	         a.release_notes, a.status, a.review_note, a.reviewed_by, a.reviewed_at,
	         a.published_at, a.created_at, a.updated_at
	         FROM apps a
	         JOIN developers d ON d.id = a.developer_id
	         WHERE a.status = 'approved'`
	args := []any{}
	if category != "" {
		query += " AND a.category = ?"
		args = append(args, category)
	}
	if queryText != "" {
		query += " AND (lower(a.name) LIKE ? OR lower(a.tagline) LIKE ? OR lower(a.description) LIKE ? OR lower(a.tags) LIKE ?)"
		like := "%" + queryText + "%"
		args = append(args, like, like, like, like)
	}
	query += " ORDER BY COALESCE(a.published_at, a.created_at) DESC, a.id DESC"
	apps, err := h.queryApps(r, query, false, args...)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "list marketplace apps", "error", err)
		writeError(w, http.StatusInternalServerError, "could not list marketplace apps")
		return
	}
	writeJSON(w, http.StatusOK, map[string][]marketplaceApp{"apps": apps})
}

func (h *Auth) GetMarketplaceApp(w http.ResponseWriter, r *http.Request) {
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
		h.logger.ErrorContext(r.Context(), "read marketplace app", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read app")
		return
	}
	writeJSON(w, http.StatusOK, map[string]marketplaceApp{"app": app})
}

func (h *Auth) GetMarketplaceFavorite(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	appID, ok := h.approvedAppIDBySlug(w, r)
	if !ok {
		return
	}
	state, err := h.readFavorite(r, userID, appID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read marketplace favorite", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read favorite")
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h *Auth) ToggleMarketplaceFavorite(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	appID, ok := h.approvedAppIDBySlug(w, r)
	if !ok {
		return
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "begin favorite tx", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update favorite")
		return
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(
		r.Context(),
		`INSERT INTO app_favorites(user_id, app_id) VALUES (?, ?)
		 ON CONFLICT(user_id, app_id) DO NOTHING`,
		userID, appID,
	)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "create favorite", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update favorite")
		return
	}
	created, err := result.RowsAffected()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read favorite result", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update favorite")
		return
	}
	favorited := created == 1
	if !favorited {
		if _, err := tx.ExecContext(r.Context(), "DELETE FROM app_favorites WHERE user_id = ? AND app_id = ?", userID, appID); err != nil {
			h.logger.ErrorContext(r.Context(), "delete favorite", "error", err)
			writeError(w, http.StatusInternalServerError, "could not update favorite")
			return
		}
	}
	if err := tx.Commit(); err != nil {
		h.logger.ErrorContext(r.Context(), "commit favorite tx", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update favorite")
		return
	}
	state, err := h.readFavorite(r, userID, appID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read updated favorite", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update favorite")
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h *Auth) approvedAppIDBySlug(w http.ResponseWriter, r *http.Request) (int64, bool) {
	slug := strings.TrimSpace(r.PathValue("slug"))
	if slug == "" {
		writeError(w, http.StatusBadRequest, "invalid app slug")
		return 0, false
	}
	var appID int64
	err := h.db.QueryRowContext(r.Context(), "SELECT id FROM apps WHERE slug = ? AND status = 'approved'", slug).Scan(&appID)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "app not found")
		return 0, false
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read favorite app", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read app")
		return 0, false
	}
	return appID, true
}

func (h *Auth) readFavorite(r *http.Request, userID, appID int64) (marketplaceFavorite, error) {
	var state marketplaceFavorite
	err := h.db.QueryRowContext(
		r.Context(),
		`SELECT EXISTS(SELECT 1 FROM app_favorites WHERE user_id = ? AND app_id = ?),
		        (SELECT COUNT(*) FROM app_favorites WHERE app_id = ?)`,
		userID, appID, appID,
	).Scan(&state.Favorited, &state.FavoriteCount)
	return state, err
}

func normalizeAppPublishRequest(request *appPublishRequest) {
	request.Name = strings.TrimSpace(request.Name)
	request.Slug = strings.ToLower(strings.TrimSpace(request.Slug))
	request.Tagline = strings.TrimSpace(request.Tagline)
	request.Description = strings.TrimSpace(request.Description)
	request.Category = strings.TrimSpace(request.Category)
	request.Currency = strings.ToUpper(strings.TrimSpace(request.Currency))
	request.IconURL = strings.TrimSpace(request.IconURL)
	request.CoverImageURL = strings.TrimSpace(request.CoverImageURL)
	request.DemoURL = strings.TrimSpace(request.DemoURL)
	request.DocsURL = strings.TrimSpace(request.DocsURL)
	request.SourceURL = strings.TrimSpace(request.SourceURL)
	request.SupportURL = strings.TrimSpace(request.SupportURL)
	request.Version = strings.TrimSpace(request.Version)
	request.ReleaseNotes = strings.TrimSpace(request.ReleaseNotes)
	cleanTags := make([]string, 0, len(request.Tags))
	seen := map[string]bool{}
	for _, tag := range request.Tags {
		tag = strings.TrimSpace(tag)
		key := strings.ToLower(tag)
		if tag != "" && !seen[key] {
			cleanTags = append(cleanTags, tag)
			seen[key] = true
		}
	}
	request.Tags = cleanTags
	if request.Currency == "" {
		request.Currency = "USD"
	}
	cleanPlans := make([]appPlanInput, 0, len(request.Plans))
	for _, plan := range request.Plans {
		plan.Name = strings.TrimSpace(plan.Name)
		plan.Description = strings.TrimSpace(plan.Description)
		plan.Currency = strings.ToUpper(strings.TrimSpace(plan.Currency))
		if plan.Currency == "" {
			plan.Currency = "USD"
		}
		cleanFeatures := make([]string, 0, len(plan.Features))
		featureSeen := map[string]bool{}
		for _, feature := range plan.Features {
			feature = strings.TrimSpace(feature)
			key := strings.ToLower(feature)
			if feature != "" && !featureSeen[key] {
				cleanFeatures = append(cleanFeatures, feature)
				featureSeen[key] = true
			}
		}
		plan.Features = cleanFeatures
		cleanPlans = append(cleanPlans, plan)
	}
	request.Plans = cleanPlans
}

func validateAppPublishRequest(request appPublishRequest) error {
	if request.Name == "" || utf8.RuneCountInString(request.Name) > 80 {
		return errors.New("name is required and must not exceed 80 characters")
	}
	if !validSlug(request.Slug) {
		return errors.New("slug must use lowercase letters, numbers, and hyphens")
	}
	if request.Tagline == "" || utf8.RuneCountInString(request.Tagline) > 140 {
		return errors.New("tagline is required and must not exceed 140 characters")
	}
	if request.Description == "" || utf8.RuneCountInString(request.Description) > 4000 {
		return errors.New("description is required and must not exceed 4000 characters")
	}
	if !appCategories[request.Category] {
		return errors.New("category is invalid")
	}
	if request.PriceCents < 0 {
		return errors.New("priceCents must be zero or greater")
	}
	if request.Currency != "USD" {
		return errors.New("currency must be USD")
	}
	if len(request.Tags) > 12 {
		return errors.New("tags must not exceed 12 items")
	}
	if request.Version == "" || utf8.RuneCountInString(request.Version) > 40 {
		return errors.New("version is required and must not exceed 40 characters")
	}
	if utf8.RuneCountInString(request.ReleaseNotes) > 2000 {
		return errors.New("releaseNotes must not exceed 2000 characters")
	}
	if len(request.Plans) > 8 {
		return errors.New("plans must not exceed 8 items")
	}
	planNames := map[string]bool{}
	for _, plan := range request.Plans {
		if plan.Name == "" || utf8.RuneCountInString(plan.Name) > 60 {
			return errors.New("plan name is required and must not exceed 60 characters")
		}
		nameKey := strings.ToLower(plan.Name)
		if planNames[nameKey] {
			return errors.New("plan names must be unique")
		}
		planNames[nameKey] = true
		if plan.PriceCents < 0 {
			return errors.New("plan priceCents must be zero or greater")
		}
		if plan.Currency != "USD" {
			return errors.New("plan currency must be USD")
		}
		if utf8.RuneCountInString(plan.Description) > 1000 {
			return errors.New("plan description must not exceed 1000 characters")
		}
		if len(plan.Features) > 20 {
			return errors.New("plan features must not exceed 20 items")
		}
		for _, feature := range plan.Features {
			if utf8.RuneCountInString(feature) > 200 {
				return errors.New("plan features must not exceed 200 characters each")
			}
		}
	}
	return nil
}

func validSlug(slug string) bool {
	if len(slug) < 3 || len(slug) > 80 || strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
		return false
	}
	previousHyphen := false
	for _, r := range slug {
		valid := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-'
		if !valid || (r == '-' && previousHyphen) {
			return false
		}
		previousHyphen = r == '-'
	}
	return true
}

func pagination(r *http.Request) (int, int) {
	limit := 50
	offset := 0
	if value, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && value > 0 && value <= 100 {
		limit = value
	}
	if value, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && value > 0 {
		offset = value
	}
	return limit, offset
}

func validAppStatus(status string) bool {
	return status == "pending_review" || status == "approved" || status == "rejected" || status == "delisted"
}

func appIDFromPath(w http.ResponseWriter, r *http.Request) (int64, bool) {
	appID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || appID < 1 {
		writeError(w, http.StatusBadRequest, "invalid app id")
		return 0, false
	}
	return appID, true
}

func (h *Auth) readApp(r *http.Request, appID int64, includeEmail bool) (marketplaceApp, error) {
	emailExpr := "''"
	if includeEmail {
		emailExpr = "u.email"
	}
	row := h.db.QueryRowContext(
		r.Context(),
		`SELECT a.id, a.developer_id, d.display_name, `+emailExpr+`, a.name, a.slug, a.tagline,
		        a.description, a.category, a.price_cents, a.currency, a.icon_url, a.cover_image_url,
		        a.demo_url, a.docs_url, a.source_url, a.support_url, a.tags, a.version,
		        a.release_notes, a.status, a.review_note, a.reviewed_by, a.reviewed_at,
		        a.published_at, a.created_at, a.updated_at
		 FROM apps a
		 JOIN developers d ON d.id = a.developer_id
		 JOIN users u ON u.id = d.user_id
		 WHERE a.id = ?`,
		appID,
	)
	return h.scanAppWithPlans(r, row)
}

func (h *Auth) readDeveloperApp(r *http.Request, appID int64, developerID int64) (marketplaceApp, error) {
	row := h.db.QueryRowContext(
		r.Context(),
		`SELECT a.id, a.developer_id, d.display_name, '', a.name, a.slug, a.tagline,
		        a.description, a.category, a.price_cents, a.currency, a.icon_url, a.cover_image_url,
		        a.demo_url, a.docs_url, a.source_url, a.support_url, a.tags, a.version,
		        a.release_notes, a.status, a.review_note, a.reviewed_by, a.reviewed_at,
		        a.published_at, a.created_at, a.updated_at
		 FROM apps a
		 JOIN developers d ON d.id = a.developer_id
		 WHERE a.id = ? AND a.developer_id = ?`,
		appID, developerID,
	)
	return h.scanAppWithPlans(r, row)
}

func (h *Auth) requireDeveloperID(w http.ResponseWriter, r *http.Request, userID int64) (int64, bool) {
	var developerID int64
	if err := h.db.QueryRowContext(r.Context(), "SELECT id FROM developers WHERE user_id = ?", userID).Scan(&developerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusForbidden, "approved developer access required")
			return 0, false
		}
		h.logger.ErrorContext(r.Context(), "read developer", "error", err)
		writeError(w, http.StatusInternalServerError, "could not verify developer access")
		return 0, false
	}
	return developerID, true
}

func (h *Auth) readAppBySlug(r *http.Request, slug string) (marketplaceApp, error) {
	row := h.db.QueryRowContext(
		r.Context(),
		`SELECT a.id, a.developer_id, d.display_name, '', a.name, a.slug, a.tagline,
		        a.description, a.category, a.price_cents, a.currency, a.icon_url, a.cover_image_url,
		        a.demo_url, a.docs_url, a.source_url, a.support_url, a.tags, a.version,
		        a.release_notes, a.status, a.review_note, a.reviewed_by, a.reviewed_at,
		        a.published_at, a.created_at, a.updated_at
		 FROM apps a
		 JOIN developers d ON d.id = a.developer_id
		 WHERE a.slug = ? AND a.status = 'approved'`,
		slug,
	)
	return h.scanAppWithPlans(r, row)
}

func (h *Auth) queryApps(r *http.Request, query string, _ bool, args ...any) ([]marketplaceApp, error) {
	rows, err := h.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		return nil, err
	}
	apps := make([]marketplaceApp, 0)
	for rows.Next() {
		app, err := scanApp(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		apps = append(apps, app)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	for i := range apps {
		plans, err := h.loadPlans(r, apps[i].ID)
		if err != nil {
			return nil, err
		}
		apps[i].Plans = plans
		if err := h.db.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM app_favorites WHERE app_id = ?", apps[i].ID).Scan(&apps[i].FavoriteCount); err != nil {
			return nil, err
		}
	}
	return apps, nil
}

func (h *Auth) scanAppWithPlans(r *http.Request, scanner appScanner) (marketplaceApp, error) {
	app, err := scanApp(scanner)
	if err != nil {
		return marketplaceApp{}, err
	}
	plans, err := h.loadPlans(r, app.ID)
	if err != nil {
		return marketplaceApp{}, err
	}
	app.Plans = plans
	if err := h.db.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM app_favorites WHERE app_id = ?", app.ID).Scan(&app.FavoriteCount); err != nil {
		return marketplaceApp{}, err
	}
	return app, nil
}

func (h *Auth) loadPlans(r *http.Request, appID int64) ([]appPlan, error) {
	rows, err := h.db.QueryContext(
		r.Context(),
		`SELECT id, app_id, name, price_cents, currency, description, features, sort_order,
		        created_at, updated_at
		 FROM app_plans
		 WHERE app_id = ?
		 ORDER BY sort_order ASC, id ASC`,
		appID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	plans := make([]appPlan, 0)
	for rows.Next() {
		var plan appPlan
		var featuresJSON string
		var createdAt, updatedAt time.Time
		if err := rows.Scan(
			&plan.ID, &plan.AppID, &plan.Name, &plan.PriceCents, &plan.Currency, &plan.Description,
			&featuresJSON, &plan.SortOrder, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(featuresJSON), &plan.Features); err != nil {
			plan.Features = []string{}
		}
		plan.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		plan.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
		plans = append(plans, plan)
	}
	return plans, rows.Err()
}

func (h *Auth) replacePlans(r *http.Request, tx *sql.Tx, appID int64, plans []appPlanInput) error {
	if _, err := tx.ExecContext(r.Context(), `DELETE FROM app_plans WHERE app_id = ?`, appID); err != nil {
		return err
	}
	for index, plan := range plans {
		featuresJSON, err := json.Marshal(plan.Features)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(
			r.Context(),
			`INSERT INTO app_plans(app_id, name, price_cents, currency, description, features, sort_order)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			appID, plan.Name, plan.PriceCents, plan.Currency, plan.Description, string(featuresJSON), index,
		); err != nil {
			return err
		}
	}
	return nil
}

type appScanner interface {
	Scan(dest ...any) error
}

func scanApp(scanner appScanner) (marketplaceApp, error) {
	var app marketplaceApp
	var tagsJSON string
	var reviewedBy sql.NullInt64
	var reviewedAt, publishedAt sql.NullTime
	var createdAt, updatedAt time.Time
	err := scanner.Scan(
		&app.ID, &app.DeveloperID, &app.DeveloperName, &app.DeveloperEmail, &app.Name,
		&app.Slug, &app.Tagline, &app.Description, &app.Category, &app.PriceCents,
		&app.Currency, &app.IconURL, &app.CoverImageURL, &app.DemoURL, &app.DocsURL,
		&app.SourceURL, &app.SupportURL, &tagsJSON, &app.Version, &app.ReleaseNotes,
		&app.Status, &app.ReviewNote, &reviewedBy, &reviewedAt, &publishedAt, &createdAt, &updatedAt,
	)
	if err != nil {
		return marketplaceApp{}, err
	}
	if err := json.Unmarshal([]byte(tagsJSON), &app.Tags); err != nil {
		app.Tags = []string{}
	}
	if reviewedBy.Valid {
		value := reviewedBy.Int64
		app.ReviewedBy = &value
	}
	if reviewedAt.Valid {
		app.ReviewedAt = reviewedAt.Time.UTC().Format(time.RFC3339)
	}
	if publishedAt.Valid {
		app.PublishedAt = publishedAt.Time.UTC().Format(time.RFC3339)
	}
	app.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	app.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
	return app, nil
}
