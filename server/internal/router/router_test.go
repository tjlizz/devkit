package router

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"devkit/server/internal/config"
	"devkit/server/internal/database"
	"devkit/server/internal/handler"
	"devkit/server/internal/middleware"
	"devkit/server/migrations"

	"golang.org/x/crypto/bcrypt"
)

func TestHealth(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	response := httptest.NewRecorder()

	New(logger).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var body handler.HealthResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != "ok" {
		t.Errorf("status = %q, want ok", body.Status)
	}
}

func TestRegisterActivateLoginAndAuthenticate(t *testing.T) {
	db, err := database.Open(":memory:", migrations.Files)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	cfg := config.Default()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := New(logger, WithAuth(db, cfg.Auth))

	registerBody := `{
		"email": "Developer@Example.com",
		"password": "strong-password",
		"displayName": "Dev Kit"
	}`
	response := performRequest(
		app,
		http.MethodPost,
		"/api/v1/auth/register",
		registerBody,
		"",
	)
	if response.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d; body = %s", response.Code, http.StatusCreated, response.Body.String())
	}

	var userID int64
	var passwordHash, avatarURL, displayName string
	var verifiedAt any
	err = db.QueryRow(
		`SELECT id, password_hash, avatar_url, display_name, verified_at
		 FROM users WHERE email = ?`,
		"developer@example.com",
	).Scan(&userID, &passwordHash, &avatarURL, &displayName, &verifiedAt)
	if err != nil {
		t.Fatalf("query registered user: %v", err)
	}
	if verifiedAt != nil {
		t.Errorf("verified_at = %v, want nil before activation", verifiedAt)
	}
	if displayName != "Dev Kit" {
		t.Errorf("display_name = %q, want Dev Kit", displayName)
	}
	if !strings.HasPrefix(avatarURL, "https://www.gravatar.com/avatar/") ||
		!strings.HasSuffix(avatarURL, "?d=identicon&s=240") {
		t.Errorf("avatar_url = %q, want 240px identicon gravatar URL", avatarURL)
	}
	cost, err := bcrypt.Cost([]byte(passwordHash))
	if err != nil {
		t.Fatalf("read bcrypt cost: %v", err)
	}
	if cost != 10 {
		t.Errorf("bcrypt cost = %d, want 10", cost)
	}

	var developerCount int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM developers WHERE user_id = ? AND display_name = ?",
		userID,
		"Dev Kit",
	).Scan(&developerCount); err != nil {
		t.Fatalf("query developer: %v", err)
	}
	if developerCount != 1 {
		t.Fatalf("developer count = %d, want 1", developerCount)
	}

	response = performRequest(
		app,
		http.MethodPost,
		"/api/v1/auth/login",
		`{"email":"developer@example.com","password":"strong-password"}`,
		"",
	)
	if response.Code != http.StatusForbidden {
		t.Fatalf("unverified login status = %d, want %d", response.Code, http.StatusForbidden)
	}

	var verificationToken string
	if err := db.QueryRow(
		"SELECT token FROM user_verifications WHERE user_id = ?",
		userID,
	).Scan(&verificationToken); err != nil {
		t.Fatalf("query verification token: %v", err)
	}
	response = performRequest(
		app,
		http.MethodGet,
		"/api/v1/auth/activate?token="+verificationToken,
		"",
		"",
	)
	if response.Code != http.StatusOK {
		t.Fatalf("activate status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	var activation struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(response.Body).Decode(&activation); err != nil {
		t.Fatalf("decode activation response: %v", err)
	}
	if activation.Status != "activated" {
		t.Errorf("activation status = %q, want activated", activation.Status)
	}

	var verificationCount int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM user_verifications WHERE token = ?",
		verificationToken,
	).Scan(&verificationCount); err != nil {
		t.Fatalf("count consumed verification token: %v", err)
	}
	if verificationCount != 0 {
		t.Errorf("verification token count = %d, want 0", verificationCount)
	}

	response = performRequest(
		app,
		http.MethodPost,
		"/api/v1/auth/login",
		`{"email":"developer@example.com","password":"strong-password"}`,
		"",
	)
	if response.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	var login struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(response.Body).Decode(&login); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if login.Token == "" {
		t.Fatal("login token is empty")
	}

	protected := middleware.JWT(cfg.Auth.JWTSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextUserID, contextEmail, ok := middleware.FromContext(r.Context())
		if !ok || contextUserID != userID || contextEmail != "developer@example.com" {
			t.Errorf(
				"auth context = (%d, %q, %t), want (%d, %q, true)",
				contextUserID,
				contextEmail,
				ok,
				userID,
				"developer@example.com",
			)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	response = performRequest(protected, http.MethodGet, "/protected", "", login.Token)
	if response.Code != http.StatusNoContent {
		t.Fatalf("protected status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func performRequest(
	app http.Handler,
	method string,
	target string,
	body string,
	token string,
) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	return response
}
