package router

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"devkit/server/internal/config"
	"devkit/server/internal/database"
	"devkit/server/migrations"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const testJWTSecret = "test-JWT-secret"

func TestRegisterCreatesUnverifiedUser(t *testing.T) {
	db, app := newAuthTestApp(t)
	response := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/register",
		`{"email":"Developer@Example.com","password":"correct-horse","displayName":"Dev One"}`,
	)
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusCreated, response.Body.String())
	}

	var body struct {
		User struct {
			ID          int64  `json:"id"`
			Email       string `json:"email"`
			DisplayName string `json:"displayName"`
			AvatarURL   string `json:"avatarUrl"`
		} `json:"user"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.User.ID < 1 || body.User.Email != "developer@example.com" {
		t.Errorf("registered user = %+v", body.User)
	}
	if body.User.DisplayName != "Dev One" {
		t.Errorf("displayName = %q, want Dev One", body.User.DisplayName)
	}
	if !strings.HasPrefix(body.User.AvatarURL, "https://www.gravatar.com/avatar/") {
		t.Errorf("avatar URL = %q, want gravatar URL", body.User.AvatarURL)
	}

	var passwordHash, avatarURL string
	var verifiedAt sql.NullTime
	if err := db.QueryRow(
		"SELECT password_hash, avatar_url, verified_at FROM users WHERE id = ?",
		body.User.ID,
	).Scan(&passwordHash, &avatarURL, &verifiedAt); err != nil {
		t.Fatalf("query registered user: %v", err)
	}
	if passwordHash == "correct-horse" {
		t.Error("password was stored in plaintext")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("correct-horse")); err != nil {
		t.Errorf("stored password hash does not match: %v", err)
	}
	if avatarURL != body.User.AvatarURL {
		t.Errorf("stored avatar URL = %q, want response value", avatarURL)
	}
	if verifiedAt.Valid {
		t.Error("newly registered user must not be verified")
	}

	var token string
	if err := db.QueryRow(
		"SELECT token FROM user_verifications WHERE user_id = ?",
		body.User.ID,
	).Scan(&token); err != nil {
		t.Fatalf("query verification token: %v", err)
	}
	if _, err := uuid.Parse(token); err != nil {
		t.Errorf("verification token = %q, want UUID: %v", token, err)
	}

	duplicate := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/register",
		`{"email":"DEVELOPER@example.com","password":"another-password","displayName":"Dev Two"}`,
	)
	if duplicate.Code != http.StatusConflict {
		t.Errorf("duplicate status = %d, want %d", duplicate.Code, http.StatusConflict)
	}
}

func TestRegisterValidatesInput(t *testing.T) {
	_, app := newAuthTestApp(t)
	tests := []struct {
		name string
		body string
	}{
		{name: "email", body: `{"email":"not-an-email","password":"long-enough","displayName":"Developer"}`},
		{name: "password", body: `{"email":"dev@example.com","password":"short","displayName":"Developer"}`},
		{name: "display name", body: `{"email":"dev@example.com","password":"long-enough","displayName":""}`},
		{name: "unknown field", body: `{"email":"dev@example.com","password":"long-enough","displayName":"Developer","admin":true}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := performJSONRequest(
				t,
				app,
				http.MethodPost,
				"/api/v1/auth/register",
				test.body,
			)
			if response.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d; body = %s", response.Code, http.StatusBadRequest, response.Body.String())
			}
		})
	}
}

func TestActivateAndLogin(t *testing.T) {
	db, app := newAuthTestApp(t)
	register := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/register",
		`{"email":"dev@example.com","password":"correct-horse","displayName":"Developer"}`,
	)
	if register.Code != http.StatusCreated {
		t.Fatalf("register status = %d; body = %s", register.Code, register.Body.String())
	}

	beforeActivation := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/login",
		`{"email":"dev@example.com","password":"correct-horse"}`,
	)
	if beforeActivation.Code != http.StatusForbidden {
		t.Errorf("unverified login status = %d, want %d", beforeActivation.Code, http.StatusForbidden)
	}

	var token string
	if err := db.QueryRow(
		`SELECT uv.token
		 FROM user_verifications uv
		 JOIN users u ON u.id = uv.user_id
		 WHERE u.email = ?`,
		"dev@example.com",
	).Scan(&token); err != nil {
		t.Fatalf("query activation token: %v", err)
	}

	activationRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/auth/activate?token="+token,
		nil,
	)
	activationResponse := httptest.NewRecorder()
	app.ServeHTTP(activationResponse, activationRequest)
	if activationResponse.Code != http.StatusOK {
		t.Fatalf("activation status = %d; body = %s", activationResponse.Code, activationResponse.Body.String())
	}

	var verified bool
	if err := db.QueryRow(
		"SELECT verified_at IS NOT NULL FROM users WHERE email = ?",
		"dev@example.com",
	).Scan(&verified); err != nil {
		t.Fatalf("query verified user: %v", err)
	}
	if !verified {
		t.Error("activated user has no verified_at value")
	}
	var verificationCount int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM user_verifications WHERE token = ?",
		token,
	).Scan(&verificationCount); err != nil {
		t.Fatalf("count consumed token: %v", err)
	}
	if verificationCount != 0 {
		t.Errorf("verification token count = %d, want 0", verificationCount)
	}

	reused := httptest.NewRecorder()
	app.ServeHTTP(reused, activationRequest)
	if reused.Code != http.StatusBadRequest {
		t.Errorf("reused token status = %d, want %d", reused.Code, http.StatusBadRequest)
	}

	login := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/login",
		`{"email":"DEV@example.com","password":"correct-horse"}`,
	)
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d; body = %s", login.Code, login.Body.String())
	}
	var loginBody struct {
		Token string `json:"token"`
		User  struct {
			Email string `json:"email"`
		} `json:"user"`
	}
	if err := json.NewDecoder(login.Body).Decode(&loginBody); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	claims := jwt.MapClaims{}
	parsed, err := jwt.ParseWithClaims(
		loginBody.Token,
		claims,
		func(_ *jwt.Token) (any, error) {
			return []byte(testJWTSecret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil || !parsed.Valid {
		t.Fatalf("parse login JWT: valid = %v, error = %v", parsed != nil && parsed.Valid, err)
	}
	if claims["email"] != "dev@example.com" || claims["user_id"] == nil {
		t.Errorf("JWT claims = %+v", claims)
	}

	wrongPassword := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/login",
		`{"email":"dev@example.com","password":"wrong-password"}`,
	)
	if wrongPassword.Code != http.StatusUnauthorized {
		t.Errorf("wrong password status = %d, want %d", wrongPassword.Code, http.StatusUnauthorized)
	}
}

func TestForgotAndResetPassword(t *testing.T) {
	db, app := newAuthTestApp(t)
	userID := insertVerifiedUser(t, db, "dev@example.com", "old-password")

	unknown := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/forgot-password",
		`{"email":"missing@example.com"}`,
	)
	if unknown.Code != http.StatusOK {
		t.Fatalf("unknown forgot status = %d, want %d; body = %s", unknown.Code, http.StatusOK, unknown.Body.String())
	}

	invalidEmail := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/forgot-password",
		`{"email":"not-an-email"}`,
	)
	if invalidEmail.Code != http.StatusBadRequest {
		t.Fatalf("invalid forgot status = %d, want %d", invalidEmail.Code, http.StatusBadRequest)
	}

	forgot := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/forgot-password",
		`{"email":"DEV@example.com"}`,
	)
	if forgot.Code != http.StatusOK {
		t.Fatalf("forgot status = %d, want %d; body = %s", forgot.Code, http.StatusOK, forgot.Body.String())
	}
	var token string
	if err := db.QueryRow(
		"SELECT token FROM password_resets WHERE user_id = ?",
		userID,
	).Scan(&token); err != nil {
		t.Fatalf("query password reset token: %v", err)
	}
	if _, err := uuid.Parse(token); err != nil {
		t.Fatalf("password reset token = %q, want UUID: %v", token, err)
	}

	shortPassword := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/reset-password",
		`{"token":"`+token+`","newPassword":"short"}`,
	)
	if shortPassword.Code != http.StatusBadRequest {
		t.Fatalf("short reset password status = %d, want %d", shortPassword.Code, http.StatusBadRequest)
	}

	reset := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/reset-password",
		`{"token":"`+token+`","newPassword":"new-password"}`,
	)
	if reset.Code != http.StatusOK {
		t.Fatalf("reset status = %d, want %d; body = %s", reset.Code, http.StatusOK, reset.Body.String())
	}

	var resetCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM password_resets WHERE token = ?", token).Scan(&resetCount); err != nil {
		t.Fatalf("count reset token: %v", err)
	}
	if resetCount != 0 {
		t.Fatalf("reset token count = %d, want 0", resetCount)
	}

	oldLogin := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/login",
		`{"email":"dev@example.com","password":"old-password"}`,
	)
	if oldLogin.Code != http.StatusUnauthorized {
		t.Fatalf("old password login status = %d, want %d", oldLogin.Code, http.StatusUnauthorized)
	}
	newLogin := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/login",
		`{"email":"dev@example.com","password":"new-password"}`,
	)
	if newLogin.Code != http.StatusOK {
		t.Fatalf("new password login status = %d, want %d; body = %s", newLogin.Code, http.StatusOK, newLogin.Body.String())
	}

	reused := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/reset-password",
		`{"token":"`+token+`","newPassword":"another-password"}`,
	)
	if reused.Code != http.StatusBadRequest {
		t.Fatalf("reused token status = %d, want %d", reused.Code, http.StatusBadRequest)
	}
}

func TestResetPasswordRejectsExpiredToken(t *testing.T) {
	db, app := newAuthTestApp(t)
	userID := insertVerifiedUser(t, db, "dev@example.com", "old-password")
	token := uuid.NewString()
	if _, err := db.Exec(
		"INSERT INTO password_resets(token, user_id, expires_at) VALUES (?, ?, datetime('now', '-1 hour'))",
		token,
		userID,
	); err != nil {
		t.Fatalf("insert expired password reset: %v", err)
	}

	response := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/reset-password",
		`{"token":"`+token+`","newPassword":"new-password"}`,
	)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expired reset status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestChangePasswordRequiresCurrentPassword(t *testing.T) {
	db, app := newAuthTestApp(t)
	insertVerifiedUser(t, db, "dev@example.com", "old-password")

	login := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/login",
		`{"email":"dev@example.com","password":"old-password"}`,
	)
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d; body = %s", login.Code, http.StatusOK, login.Body.String())
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(login.Body).Decode(&loginBody); err != nil {
		t.Fatalf("decode login: %v", err)
	}

	unauthenticated := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/change-password",
		`{"oldPassword":"old-password","newPassword":"new-password"}`,
	)
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated change status = %d, want %d", unauthenticated.Code, http.StatusUnauthorized)
	}

	wrongCurrent := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/change-password",
		`{"oldPassword":"wrong-password","newPassword":"new-password"}`,
		loginBody.Token,
	)
	if wrongCurrent.Code != http.StatusUnauthorized {
		t.Fatalf("wrong current status = %d, want %d", wrongCurrent.Code, http.StatusUnauthorized)
	}

	shortPassword := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/change-password",
		`{"oldPassword":"old-password","newPassword":"short"}`,
		loginBody.Token,
	)
	if shortPassword.Code != http.StatusBadRequest {
		t.Fatalf("short new password status = %d, want %d", shortPassword.Code, http.StatusBadRequest)
	}

	changed := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/change-password",
		`{"oldPassword":"old-password","newPassword":"new-password"}`,
		loginBody.Token,
	)
	if changed.Code != http.StatusOK {
		t.Fatalf("change status = %d, want %d; body = %s", changed.Code, http.StatusOK, changed.Body.String())
	}

	oldLogin := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/login",
		`{"email":"dev@example.com","password":"old-password"}`,
	)
	if oldLogin.Code != http.StatusUnauthorized {
		t.Fatalf("old password login status = %d, want %d", oldLogin.Code, http.StatusUnauthorized)
	}
	newLogin := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/login",
		`{"email":"dev@example.com","password":"new-password"}`,
	)
	if newLogin.Code != http.StatusOK {
		t.Fatalf("new password login status = %d, want %d; body = %s", newLogin.Code, http.StatusOK, newLogin.Body.String())
	}
}

func newAuthTestApp(t *testing.T) (*sql.DB, http.Handler) {
	t.Helper()
	db, err := database.Open(":memory:", migrations.Files)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Default().Auth
	cfg.JWTSecret = testJWTSecret
	return db, New(logger, WithAuth(db, cfg))
}

func insertVerifiedUser(t *testing.T, db *sql.DB, email string, password string) int64 {
	t.Helper()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	result, err := db.Exec(
		`INSERT INTO users(email, password_hash, avatar_url, display_name, verified_at)
		 VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		email,
		string(passwordHash),
		defaultTestAvatar(email),
		"Developer",
	)
	if err != nil {
		t.Fatalf("insert verified user: %v", err)
	}
	userID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read user ID: %v", err)
	}
	return userID
}

func defaultTestAvatar(email string) string {
	return "https://example.com/avatar/" + email
}

func performJSONRequest(
	t *testing.T,
	app http.Handler,
	method string,
	target string,
	body string,
) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	return response
}

func performJSONRequestWithToken(
	t *testing.T,
	app http.Handler,
	method string,
	target string,
	body string,
	token string,
) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	return response
}
