package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// DeveloperProfile is the public profile of a developer, plus a summary of
// their work. Apps only includes approved marketplace apps.
type DeveloperProfile struct {
	ID             int64            `json:"id"`
	Username       string           `json:"username"`
	Name           string           `json:"name"`
	AvatarURL      string           `json:"avatarUrl"`
	Bio            string           `json:"bio"`
	Location       string           `json:"location"`
	Website        string           `json:"website"`
	Email          string           `json:"email"`
	JoinedAt       string           `json:"joinedAt"`
	PublishedCount int              `json:"publishedCount"`
	Apps           []marketplaceApp `json:"apps"`
}

// developerSlug builds a stable, URL-safe profile slug for a developer from
// their id. Because display names are not unique, we key public profiles by
// "developer-<id>", matching the href the marketplace already emits.
func developerSlug(id int64) string {
	return "developer-" + strconv.FormatInt(id, 10)
}

// GetDeveloperProfile returns a developer's public profile and their approved
// apps. Public endpoint; no authentication required.
func (h *Auth) GetDeveloperProfile(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(r.PathValue("slug"))
	if slug == "" {
		writeError(w, http.StatusBadRequest, "invalid developer")
		return
	}

	// Resolve slug -> developer id.
	var developerID int64
	if rest, ok := strings.CutPrefix(slug, "developer-"); ok {
		id, err := strconv.ParseInt(rest, 10, 64)
		if err != nil || id <= 0 {
			writeError(w, http.StatusNotFound, "developer not found")
			return
		}
		developerID = id
	} else {
		// Fall back to display-name slug lookup for legacy/SEO links.
		err := h.db.QueryRowContext(
			r.Context(),
			`SELECT id FROM developers WHERE lower(display_name) = lower(?)`,
			strings.ReplaceAll(slug, "-", " "),
		).Scan(&developerID)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "developer not found")
			return
		}
		if err != nil {
			h.logger.ErrorContext(r.Context(), "lookup developer by name", "error", err)
			writeError(w, http.StatusInternalServerError, "could not read developer")
			return
		}
	}

	profile, err := h.readDeveloperProfile(r, developerID)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "developer not found")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read developer profile", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read developer")
		return
	}
	writeJSON(w, http.StatusOK, map[string]DeveloperProfile{"developer": profile})
}

func (h *Auth) readDeveloperProfile(r *http.Request, developerID int64) (DeveloperProfile, error) {
	var profile DeveloperProfile
	var joinedAt time.Time
	err := h.db.QueryRowContext(
		r.Context(),
		`SELECT d.id, u.email, d.display_name, COALESCE(u.avatar_url, ''), d.bio, d.location, d.website, u.created_at
		 FROM developers d
		 JOIN users u ON u.id = d.user_id
		 WHERE d.id = ?`,
		developerID,
	).Scan(&profile.ID, &profile.Email, &profile.Name, &profile.AvatarURL,
		&profile.Bio, &profile.Location, &profile.Website, &joinedAt)
	if err != nil {
		return DeveloperProfile{}, err
	}
	profile.Username = developerSlug(profile.ID)
	profile.JoinedAt = joinedAt.UTC().Format(time.RFC3339)

	// Approved apps only, with plans, favorite counts, and review aggregates.
	apps, err := h.queryApps(r,
		`SELECT a.id, a.developer_id, d.display_name, '', a.name, a.slug, a.tagline,
		         a.description, a.category, a.price_cents, a.currency, a.icon_url, a.cover_image_url,
		         a.demo_url, a.docs_url, a.source_url, a.support_url, a.tags, a.version,
		         a.release_notes, a.status, a.review_note, a.reviewed_by, a.reviewed_at,
		         a.published_at, a.created_at, a.updated_at
		 FROM apps a
		 JOIN developers d ON d.id = a.developer_id
		 WHERE a.developer_id = ? AND a.status = 'approved'
		 ORDER BY COALESCE(a.published_at, a.created_at) DESC, a.id DESC`,
		false, developerID,
	)
	if err != nil {
		return DeveloperProfile{}, err
	}
	profile.Apps = apps
	profile.PublishedCount = len(apps)
	return profile, nil
}
