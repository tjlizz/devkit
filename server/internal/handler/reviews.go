package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"devkit/server/internal/middleware"
)

// Review is a buyer review of a marketplace app. VerifiedPurchase is true when
// the reviewer holds an active entitlement for the app.
type Review struct {
	ID               int64  `json:"id"`
	AppID            int64  `json:"appId"`
	AppSlug          string `json:"appSlug"`
	BuyerID          int64  `json:"buyerId"`
	BuyerName        string `json:"buyerName"`
	Rating           int    `json:"rating"`
	Comment          string `json:"comment"`
	VerifiedPurchase bool   `json:"verifiedPurchase"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

type reviewRequest struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

// loadReviewAggregates fills the average rating and active review count for an
// app. When no active reviews exist the rating is 0 and the count is 0.
func (h *Auth) loadReviewAggregates(r *http.Request, appID int64, ratingOut *float64, countOut *int) error {
	var count int
	var sum sql.NullFloat64
	err := h.db.QueryRowContext(
		r.Context(),
		`SELECT COUNT(*), SUM(rating) FROM app_reviews WHERE app_id = ? AND status = 'active'`,
		appID,
	).Scan(&count, &sum)
	if err != nil {
		return err
	}
	*countOut = count
	if count > 0 && sum.Valid {
		*ratingOut = sum.Float64 / float64(count)
	}
	return nil
}

// ListAppReviews returns all active reviews for an approved marketplace app,
// newest first. Public endpoint; no authentication required.
func (h *Auth) ListAppReviews(w http.ResponseWriter, r *http.Request) {
	appID, ok := h.approvedAppIDBySlug(w, r)
	if !ok {
		return
	}
	rows, err := h.db.QueryContext(
		r.Context(),
		`SELECT r.id, r.app_id, a.slug, r.buyer_id, u.display_name, r.rating, r.comment,
		        EXISTS(SELECT 1 FROM entitlements e
		               WHERE e.app_id = r.app_id AND e.buyer_id = r.buyer_id AND e.status = 'active'),
		        r.created_at, r.updated_at
		 FROM app_reviews r
		 JOIN apps a ON a.id = r.app_id
		 JOIN users u ON u.id = r.buyer_id
		 WHERE r.app_id = ? AND r.status = 'active'
		 ORDER BY r.created_at DESC, r.id DESC`,
		appID,
	)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "list app reviews", "error", err)
		writeError(w, http.StatusInternalServerError, "could not list reviews")
		return
	}
	defer rows.Close()
	reviews := make([]Review, 0)
	for rows.Next() {
		var review Review
		var createdAt, updatedAt time.Time
		if err := rows.Scan(
			&review.ID, &review.AppID, &review.AppSlug, &review.BuyerID, &review.BuyerName,
			&review.Rating, &review.Comment, &review.VerifiedPurchase, &createdAt, &updatedAt,
		); err != nil {
			h.logger.ErrorContext(r.Context(), "scan app review", "error", err)
			writeError(w, http.StatusInternalServerError, "could not list reviews")
			return
		}
		review.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		review.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		h.logger.ErrorContext(r.Context(), "iterate app reviews", "error", err)
		writeError(w, http.StatusInternalServerError, "could not list reviews")
		return
	}
	writeJSON(w, http.StatusOK, map[string][]Review{"reviews": reviews})
}

// MyAppReview returns the current user's review of an approved app, or 404 if
// they have not reviewed it. Used by the frontend to pre-fill the edit form.
func (h *Auth) MyAppReview(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	appID, ok := h.approvedAppIDBySlug(w, r)
	if !ok {
		return
	}
	review, err := h.readReview(r, appID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "review not found")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read my review", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read review")
		return
	}
	writeJSON(w, http.StatusOK, map[string]Review{"review": review})
}

// CreateOrUpdateReview lets a verified buyer with an active entitlement create
// or update their review of an approved app. Free apps require a paid order
// that was confirmed; the entitlement check enforces verified purchase.
func (h *Auth) CreateOrUpdateReview(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	appID, ok := h.approvedAppIDBySlug(w, r)
	if !ok {
		return
	}

	var request reviewRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if request.Rating < 1 || request.Rating > 5 {
		writeError(w, http.StatusBadRequest, "rating must be between 1 and 5")
		return
	}
	request.Comment = strings.TrimSpace(request.Comment)
	if utf8.RuneCountInString(request.Comment) > 2000 {
		writeError(w, http.StatusBadRequest, "comment must be 2000 characters or fewer")
		return
	}

	// Only a verified buyer with an active entitlement may review.
	var entitlementStatus string
	err := h.db.QueryRowContext(
		r.Context(),
		`SELECT status FROM entitlements WHERE buyer_id = ? AND app_id = ? AND status = 'active'`,
		userID, appID,
	).Scan(&entitlementStatus)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusForbidden, "only verified buyers with an active purchase can review this app")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "check review entitlement", "error", err)
		writeError(w, http.StatusInternalServerError, "could not verify purchase")
		return
	}

	// Upsert: one review per buyer per app.
	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "begin review tx", "error", err)
		writeError(w, http.StatusInternalServerError, "could not save review")
		return
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(
		r.Context(),
		`INSERT INTO app_reviews(app_id, buyer_id, rating, comment, status)
		 VALUES (?, ?, ?, ?, 'active')
		 ON CONFLICT(app_id, buyer_id)
		 DO UPDATE SET rating = excluded.rating, comment = excluded.comment,
		               status = 'active', updated_at = CURRENT_TIMESTAMP`,
		appID, userID, request.Rating, request.Comment,
	)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "upsert review", "error", err)
		writeError(w, http.StatusInternalServerError, "could not save review")
		return
	}
	review, err := h.readReviewTx(r, tx, appID, userID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read saved review", "error", err)
		writeError(w, http.StatusInternalServerError, "could not save review")
		return
	}
	if err := tx.Commit(); err != nil {
		h.logger.ErrorContext(r.Context(), "commit review tx", "error", err)
		writeError(w, http.StatusInternalServerError, "could not save review")
		return
	}
	writeJSON(w, http.StatusOK, map[string]Review{"review": review})
}

// DeleteReview lets a buyer remove their own review of an app (soft delete).
func (h *Auth) DeleteReview(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	appID, ok := h.approvedAppIDBySlug(w, r)
	if !ok {
		return
	}
	result, err := h.db.ExecContext(
		r.Context(),
		`UPDATE app_reviews SET status = 'hidden', updated_at = CURRENT_TIMESTAMP
		 WHERE app_id = ? AND buyer_id = ? AND status = 'active'`,
		appID, userID,
	)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "delete review", "error", err)
		writeError(w, http.StatusInternalServerError, "could not delete review")
		return
	}
	affected, err := result.RowsAffected()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "delete review rows", "error", err)
		writeError(w, http.StatusInternalServerError, "could not delete review")
		return
	}
	if affected == 0 {
		writeError(w, http.StatusNotFound, "review not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *Auth) readReview(r *http.Request, appID, buyerID int64) (Review, error) {
	return h.scanReview(h.db.QueryRowContext(
		r.Context(),
		`SELECT r.id, r.app_id, a.slug, r.buyer_id, u.display_name, r.rating, r.comment,
		        EXISTS(SELECT 1 FROM entitlements e
		               WHERE e.app_id = r.app_id AND e.buyer_id = r.buyer_id AND e.status = 'active'),
		        r.created_at, r.updated_at
		 FROM app_reviews r
		 JOIN apps a ON a.id = r.app_id
		 JOIN users u ON u.id = r.buyer_id
		 WHERE r.app_id = ? AND r.buyer_id = ? AND r.status = 'active'`,
		appID, buyerID,
	))
}

func (h *Auth) readReviewTx(r *http.Request, tx *sql.Tx, appID, buyerID int64) (Review, error) {
	return h.scanReview(tx.QueryRowContext(
		r.Context(),
		`SELECT r.id, r.app_id, a.slug, r.buyer_id, u.display_name, r.rating, r.comment,
		        EXISTS(SELECT 1 FROM entitlements e
		               WHERE e.app_id = r.app_id AND e.buyer_id = r.buyer_id AND e.status = 'active'),
		        r.created_at, r.updated_at
		 FROM app_reviews r
		 JOIN apps a ON a.id = r.app_id
		 JOIN users u ON u.id = r.buyer_id
		 WHERE r.app_id = ? AND r.buyer_id = ? AND r.status = 'active'`,
		appID, buyerID,
	))
}

func (h *Auth) scanReview(scanner appScanner) (Review, error) {
	var review Review
	var createdAt, updatedAt time.Time
	err := scanner.Scan(
		&review.ID, &review.AppID, &review.AppSlug, &review.BuyerID, &review.BuyerName,
		&review.Rating, &review.Comment, &review.VerifiedPurchase, &createdAt, &updatedAt,
	)
	if err != nil {
		return Review{}, err
	}
	review.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	review.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
	return review, nil
}
