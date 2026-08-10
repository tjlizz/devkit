package router

import (
	"encoding/json"
	"net/http"
	"testing"
)

// TestListMyFavorites verifies a buyer can list their favorited marketplace
// apps, that only approved apps they favorited are returned, that the list is
// scoped to the authenticated user, and that unauthenticated access is rejected.
func TestListMyFavorites(t *testing.T) {
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

	// Publish and approve two apps.
	publishAndApprove(t, app, developerToken, adminToken, "Favorite One", "favorite-one")
	publishAndApprove(t, app, developerToken, adminToken, "Favorite Two", "favorite-two")

	// --- Unauthenticated list is rejected. ---
	unauth := performJSONRequest(t, app, http.MethodGet, "/api/v1/me/favorites", "")
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated favorites status = %d, want %d", unauth.Code, http.StatusUnauthorized)
	}

	// --- Empty list for a user with no favorites. ---
	empty := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/me/favorites", "", buyerToken)
	if empty.Code != http.StatusOK {
		t.Fatalf("empty favorites status = %d, want %d; body = %s", empty.Code, http.StatusOK, empty.Body.String())
	}
	var emptyBody struct {
		Apps []struct {
			Slug string `json:"slug"`
		} `json:"apps"`
	}
	if err := json.NewDecoder(empty.Body).Decode(&emptyBody); err != nil {
		t.Fatalf("decode empty favorites: %v", err)
	}
	if len(emptyBody.Apps) != 0 {
		t.Fatalf("empty favorites apps = %d, want 0", len(emptyBody.Apps))
	}

	// --- Buyer favorites one app. ---
	favorite := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/marketplace/apps/favorite-one/favorite",
		"",
		buyerToken,
	)
	if favorite.Code != http.StatusOK {
		t.Fatalf("favorite status = %d, want %d; body = %s", favorite.Code, http.StatusOK, favorite.Body.String())
	}

	// --- Another user favorites a different app; must not leak into buyer list. ---
	favoriteOther := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/marketplace/apps/favorite-two/favorite",
		"",
		otherToken,
	)
	if favoriteOther.Code != http.StatusOK {
		t.Fatalf("other favorite status = %d, want %d; body = %s", favoriteOther.Code, http.StatusOK, favoriteOther.Body.String())
	}

	// --- Buyer's list contains only their favorited app. ---
	list := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/me/favorites", "", buyerToken)
	if list.Code != http.StatusOK {
		t.Fatalf("favorites status = %d, want %d; body = %s", list.Code, http.StatusOK, list.Body.String())
	}
	var listBody struct {
		Apps []struct {
			Slug          string `json:"slug"`
			Name          string `json:"name"`
			FavoriteCount int    `json:"favoriteCount"`
		} `json:"apps"`
	}
	if err := json.NewDecoder(list.Body).Decode(&listBody); err != nil {
		t.Fatalf("decode favorites: %v", err)
	}
	if len(listBody.Apps) != 1 {
		t.Fatalf("favorites apps = %d, want 1; body = %s", len(listBody.Apps), list.Body.String())
	}
	if listBody.Apps[0].Slug != "favorite-one" {
		t.Fatalf("favorites[0].slug = %q, want favorite-one", listBody.Apps[0].Slug)
	}
	if listBody.Apps[0].Name != "Favorite One" {
		t.Fatalf("favorites[0].name = %q, want Favorite One", listBody.Apps[0].Name)
	}
	if listBody.Apps[0].FavoriteCount != 1 {
		t.Fatalf("favorites[0].favoriteCount = %d, want 1", listBody.Apps[0].FavoriteCount)
	}

	// --- Other user's list contains only their app. ---
	otherList := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/me/favorites", "", otherToken)
	if otherList.Code != http.StatusOK {
		t.Fatalf("other favorites status = %d, want %d", otherList.Code, http.StatusOK)
	}
	var otherListBody struct {
		Apps []struct {
			Slug string `json:"slug"`
		} `json:"apps"`
	}
	if err := json.NewDecoder(otherList.Body).Decode(&otherListBody); err != nil {
		t.Fatalf("decode other favorites: %v", err)
	}
	if len(otherListBody.Apps) != 1 || otherListBody.Apps[0].Slug != "favorite-two" {
		t.Fatalf("other favorites = %+v, want [favorite-two]", otherListBody.Apps)
	}

	// --- Unfavoriting removes the app from the list. ---
	unfavorite := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/marketplace/apps/favorite-one/favorite",
		"",
		buyerToken,
	)
	if unfavorite.Code != http.StatusOK {
		t.Fatalf("unfavorite status = %d, want %d", unfavorite.Code, http.StatusOK)
	}
	after := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/me/favorites", "", buyerToken)
	if after.Code != http.StatusOK {
		t.Fatalf("favorites after unfavorite status = %d, want %d", after.Code, http.StatusOK)
	}
	var afterBody struct {
		Apps []struct{} `json:"apps"`
	}
	if err := json.NewDecoder(after.Body).Decode(&afterBody); err != nil {
		t.Fatalf("decode favorites after unfavorite: %v", err)
	}
	if len(afterBody.Apps) != 0 {
		t.Fatalf("favorites after unfavorite = %d, want 0", len(afterBody.Apps))
	}
}