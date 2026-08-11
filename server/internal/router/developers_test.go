package router

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

// TestDeveloperProfile verifies the public developer profile endpoint: it
// returns the developer's profile fields, only approved apps, and resolves the
// marketplace "developer-<id>" slug used across product cards.
func TestDeveloperProfile(t *testing.T) {
	db, app := newAuthTestApp(t)
	devID := insertVerifiedUser(t, db, "dev@example.com", "old-password")
	insertVerifiedUser(t, db, "buyer@example.com", "old-password")
	adminID := insertVerifiedUser(t, db, "admin@example.com", "old-password")
	if _, err := db.Exec("UPDATE users SET role = 'admin' WHERE id = ?", adminID); err != nil {
		t.Fatalf("mark admin: %v", err)
	}
	if _, err := db.Exec(
		"INSERT INTO developers(user_id, display_name, bio, location, website) VALUES (?, ?, ?, ?, ?)",
		devID, "Dev Studio", "Making calm tools.", "Vancouver", "https://dev.build",
	); err != nil {
		t.Fatalf("insert developer: %v", err)
	}

	developerToken := loginToken(t, app, "dev@example.com", "old-password")
	adminToken := loginToken(t, app, "admin@example.com", "old-password")

	// Publish two apps; approve one, leave one pending.
	publishAndApprove(t, app, developerToken, adminToken, "Approved App", "approved-app")
	pending := performJSONRequestWithToken(
		t, app, http.MethodPost, "/api/v1/apps",
		`{"name":"Pending App","slug":"pending-app","tagline":"T.","description":"D.","category":"developer-tools","priceCents":0,"currency":"USD","iconUrl":"","coverImageUrl":"","demoUrl":"","docsUrl":"","sourceUrl":"","supportUrl":"","tags":[],"version":"1.0.0","releaseNotes":"R."}`,
		developerToken,
	)
	if pending.Code != http.StatusCreated {
		t.Fatalf("publish pending status = %d, want %d", pending.Code, http.StatusCreated)
	}

	slug := "developer-" + strconv.FormatInt(devID, 10)

	// --- Public read returns profile + only the approved app. ---
	resp := performJSONRequest(t, app, http.MethodGet, "/api/v1/developers/"+slug, "")
	if resp.Code != http.StatusOK {
		t.Fatalf("profile status = %d, want %d; body = %s", resp.Code, http.StatusOK, resp.Body.String())
	}
	var body struct {
		Developer struct {
			ID             int64  `json:"id"`
			Username       string `json:"username"`
			Name           string `json:"name"`
			Bio            string `json:"bio"`
			Location       string `json:"location"`
			Website        string `json:"website"`
			PublishedCount int    `json:"publishedCount"`
			Apps           []struct {
				Slug string `json:"slug"`
			} `json:"apps"`
		} `json:"developer"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if body.Developer.ID != devID {
		t.Fatalf("developer id = %d, want %d", body.Developer.ID, devID)
	}
	if body.Developer.Username != slug {
		t.Fatalf("username = %q, want %q", body.Developer.Username, slug)
	}
	if body.Developer.Name != "Dev Studio" || body.Developer.Bio != "Making calm tools." ||
		body.Developer.Location != "Vancouver" || body.Developer.Website != "https://dev.build" {
		t.Fatalf("profile fields = %+v", body.Developer)
	}
	if body.Developer.PublishedCount != 1 || len(body.Developer.Apps) != 1 {
		t.Fatalf("published count = %d (%d apps), want 1 approved app",
			body.Developer.PublishedCount, len(body.Developer.Apps))
	}
	if body.Developer.Apps[0].Slug != "approved-app" {
		t.Fatalf("app slug = %q, want approved-app", body.Developer.Apps[0].Slug)
	}

	// --- Unknown (non-numeric) slug is a 404. ---
	missing := performJSONRequest(t, app, http.MethodGet, "/api/v1/developers/developer-999999", "")
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing profile status = %d, want %d", missing.Code, http.StatusNotFound)
	}
	badSlug := performJSONRequest(t, app, http.MethodGet, "/api/v1/developers/not-a-real-slug", "")
	if badSlug.Code != http.StatusNotFound {
		t.Fatalf("bad slug status = %d, want %d", badSlug.Code, http.StatusNotFound)
	}
}
