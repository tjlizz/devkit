package router

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestAuthenticatedUserCanUpdateAvatar(t *testing.T) {
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

	unauthenticated := performMultipartAvatarRequest(
		t,
		app,
		"",
		validPNGAvatar(),
	)
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated avatar status = %d, want %d", unauthenticated.Code, http.StatusUnauthorized)
	}

	invalid := performMultipartAvatarRequest(
		t,
		app,
		loginBody.Token,
		[]byte("not an image"),
	)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid avatar status = %d, want %d; body = %s", invalid.Code, http.StatusBadRequest, invalid.Body.String())
	}

	custom := performMultipartAvatarRequest(
		t,
		app,
		loginBody.Token,
		validPNGAvatar(),
	)
	if custom.Code != http.StatusOK {
		t.Fatalf("custom avatar status = %d, want %d; body = %s", custom.Code, http.StatusOK, custom.Body.String())
	}
	var customBody struct {
		User struct {
			AvatarURL string `json:"avatarUrl"`
		} `json:"user"`
	}
	if err := json.NewDecoder(custom.Body).Decode(&customBody); err != nil {
		t.Fatalf("decode custom avatar: %v", err)
	}
	if !strings.HasPrefix(customBody.User.AvatarURL, "/api/v1/uploads/avatars/user-") ||
		!strings.HasSuffix(customBody.User.AvatarURL, ".png") {
		t.Fatalf("avatarUrl = %q, want uploaded avatar path", customBody.User.AvatarURL)
	}

	var storedAvatarURL string
	if err := db.QueryRow("SELECT avatar_url FROM users WHERE email = ?", "dev@example.com").Scan(&storedAvatarURL); err != nil {
		t.Fatalf("query avatar URL: %v", err)
	}
	if storedAvatarURL != customBody.User.AvatarURL {
		t.Fatalf("stored avatarUrl = %q, want response URL", storedAvatarURL)
	}

	me := performJSONRequestWithToken(
		t,
		app,
		http.MethodGet,
		"/api/v1/auth/me",
		``,
		loginBody.Token,
	)
	if me.Code != http.StatusOK {
		t.Fatalf("me status = %d, want %d; body = %s", me.Code, http.StatusOK, me.Body.String())
	}
	var meBody struct {
		User struct {
			Email     string `json:"email"`
			AvatarURL string `json:"avatarUrl"`
		} `json:"user"`
	}
	if err := json.NewDecoder(me.Body).Decode(&meBody); err != nil {
		t.Fatalf("decode me: %v", err)
	}
	if meBody.User.Email != "dev@example.com" || meBody.User.AvatarURL != customBody.User.AvatarURL {
		t.Fatalf("me user = %+v, want current updated user", meBody.User)
	}

	reset := performJSONRequestWithToken(
		t,
		app,
		http.MethodDelete,
		"/api/v1/auth/me/avatar",
		``,
		loginBody.Token,
	)
	if reset.Code != http.StatusOK {
		t.Fatalf("reset avatar status = %d, want %d; body = %s", reset.Code, http.StatusOK, reset.Body.String())
	}
	var resetBody struct {
		User struct {
			AvatarURL string `json:"avatarUrl"`
		} `json:"user"`
	}
	if err := json.NewDecoder(reset.Body).Decode(&resetBody); err != nil {
		t.Fatalf("decode reset avatar: %v", err)
	}
	if !strings.HasPrefix(resetBody.User.AvatarURL, "https://www.gravatar.com/avatar/") {
		t.Fatalf("reset avatarUrl = %q, want gravatar URL", resetBody.User.AvatarURL)
	}
}

func TestDeveloperApplicationApprovalFlow(t *testing.T) {
	db, app := newAuthTestApp(t)
	insertVerifiedUser(t, db, "dev@example.com", "old-password")
	adminID := insertVerifiedUser(t, db, "admin@example.com", "old-password")
	if _, err := db.Exec("UPDATE users SET role = 'admin' WHERE id = ?", adminID); err != nil {
		t.Fatalf("mark admin user: %v", err)
	}

	userToken := loginToken(t, app, "dev@example.com", "old-password")
	adminToken := loginToken(t, app, "admin@example.com", "old-password")
	applicationBody := `{"displayName":"Dev Studio","profileUrl":"https://example.com/dev","reason":"I build production developer tools for teams."}`

	submitted := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/upgrade-to-developer",
		applicationBody,
		userToken,
	)
	if submitted.Code != http.StatusCreated {
		t.Fatalf("submit application status = %d, want %d; body = %s", submitted.Code, http.StatusCreated, submitted.Body.String())
	}
	var submittedBody struct {
		Application struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"application"`
	}
	if err := json.NewDecoder(submitted.Body).Decode(&submittedBody); err != nil {
		t.Fatalf("decode submitted application: %v", err)
	}
	if submittedBody.Application.ID < 1 || submittedBody.Application.Status != "pending" {
		t.Fatalf("submitted application = %+v, want pending application", submittedBody.Application)
	}

	var developerCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM developers WHERE user_id = (SELECT id FROM users WHERE email = ?)", "dev@example.com").Scan(&developerCount); err != nil {
		t.Fatalf("count developers: %v", err)
	}
	if developerCount != 0 {
		t.Fatalf("developer count after submit = %d, want 0", developerCount)
	}

	duplicate := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/upgrade-to-developer",
		applicationBody,
		userToken,
	)
	if duplicate.Code != http.StatusConflict {
		t.Fatalf("duplicate application status = %d, want %d", duplicate.Code, http.StatusConflict)
	}

	me := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/auth/me", "", userToken)
	if me.Code != http.StatusOK {
		t.Fatalf("me status = %d, want %d; body = %s", me.Code, http.StatusOK, me.Body.String())
	}
	var meBody struct {
		User struct {
			Role                       string `json:"role"`
			IsDeveloper                bool   `json:"isDeveloper"`
			DeveloperApplicationStatus string `json:"developerApplicationStatus"`
		} `json:"user"`
	}
	if err := json.NewDecoder(me.Body).Decode(&meBody); err != nil {
		t.Fatalf("decode me: %v", err)
	}
	if meBody.User.Role != "user" || meBody.User.IsDeveloper || meBody.User.DeveloperApplicationStatus != "pending" {
		t.Fatalf("me user = %+v, want user with pending application", meBody.User)
	}

	nonAdminList := performJSONRequestWithToken(
		t,
		app,
		http.MethodGet,
		"/api/v1/admin/developer-applications",
		"",
		userToken,
	)
	if nonAdminList.Code != http.StatusForbidden {
		t.Fatalf("non-admin list status = %d, want %d", nonAdminList.Code, http.StatusForbidden)
	}

	nonAdminApprove := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/admin/developer-applications/"+strconv.FormatInt(submittedBody.Application.ID, 10)+"/approve",
		`{"reviewNote":"Looks good."}`,
		userToken,
	)
	if nonAdminApprove.Code != http.StatusForbidden {
		t.Fatalf("non-admin approve status = %d, want %d", nonAdminApprove.Code, http.StatusForbidden)
	}

	list := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/admin/developer-applications", "", adminToken)
	if list.Code != http.StatusOK {
		t.Fatalf("admin list status = %d, want %d; body = %s", list.Code, http.StatusOK, list.Body.String())
	}
	var listBody struct {
		Applications []struct {
			ID     int64  `json:"id"`
			Email  string `json:"email"`
			Status string `json:"status"`
		} `json:"applications"`
	}
	if err := json.NewDecoder(list.Body).Decode(&listBody); err != nil {
		t.Fatalf("decode admin list: %v", err)
	}
	if len(listBody.Applications) != 1 || listBody.Applications[0].Email != "dev@example.com" || listBody.Applications[0].Status != "pending" {
		t.Fatalf("applications = %+v, want one pending dev@example.com application", listBody.Applications)
	}

	approved := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/admin/developer-applications/"+strconv.FormatInt(submittedBody.Application.ID, 10)+"/approve",
		`{"reviewNote":"Looks good."}`,
		adminToken,
	)
	if approved.Code != http.StatusOK {
		t.Fatalf("approve status = %d, want %d; body = %s", approved.Code, http.StatusOK, approved.Body.String())
	}
	var approvedBody struct {
		Application struct {
			Status     string `json:"status"`
			ReviewNote string `json:"reviewNote"`
		} `json:"application"`
	}
	if err := json.NewDecoder(approved.Body).Decode(&approvedBody); err != nil {
		t.Fatalf("decode approved application: %v", err)
	}
	if approvedBody.Application.Status != "approved" || approvedBody.Application.ReviewNote != "Looks good." {
		t.Fatalf("approved application = %+v, want approved with note", approvedBody.Application)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM developers WHERE display_name = ?", "Dev Studio").Scan(&developerCount); err != nil {
		t.Fatalf("count approved developers: %v", err)
	}
	if developerCount != 1 {
		t.Fatalf("developer count after approval = %d, want 1", developerCount)
	}

	reapprove := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/admin/developer-applications/"+strconv.FormatInt(submittedBody.Application.ID, 10)+"/approve",
		`{"reviewNote":"Again."}`,
		adminToken,
	)
	if reapprove.Code != http.StatusConflict {
		t.Fatalf("reapprove status = %d, want %d", reapprove.Code, http.StatusConflict)
	}
}

func TestDeveloperApplicationRejectionFlow(t *testing.T) {
	db, app := newAuthTestApp(t)
	insertVerifiedUser(t, db, "dev@example.com", "old-password")
	adminID := insertVerifiedUser(t, db, "admin@example.com", "old-password")
	if _, err := db.Exec("UPDATE users SET role = 'admin' WHERE id = ?", adminID); err != nil {
		t.Fatalf("mark admin user: %v", err)
	}

	userToken := loginToken(t, app, "dev@example.com", "old-password")
	adminToken := loginToken(t, app, "admin@example.com", "old-password")
	submitted := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/upgrade-to-developer",
		`{"displayName":"Dev Studio","profileUrl":"","reason":"I build production developer tools for teams."}`,
		userToken,
	)
	if submitted.Code != http.StatusCreated {
		t.Fatalf("submit application status = %d, want %d; body = %s", submitted.Code, http.StatusCreated, submitted.Body.String())
	}
	var submittedBody struct {
		Application struct {
			ID int64 `json:"id"`
		} `json:"application"`
	}
	if err := json.NewDecoder(submitted.Body).Decode(&submittedBody); err != nil {
		t.Fatalf("decode submitted application: %v", err)
	}

	missingNote := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/admin/developer-applications/"+strconv.FormatInt(submittedBody.Application.ID, 10)+"/reject",
		`{"reviewNote":""}`,
		adminToken,
	)
	if missingNote.Code != http.StatusBadRequest {
		t.Fatalf("missing note rejection status = %d, want %d", missingNote.Code, http.StatusBadRequest)
	}

	rejected := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/admin/developer-applications/"+strconv.FormatInt(submittedBody.Application.ID, 10)+"/reject",
		`{"reviewNote":"Please add more production references."}`,
		adminToken,
	)
	if rejected.Code != http.StatusOK {
		t.Fatalf("reject status = %d, want %d; body = %s", rejected.Code, http.StatusOK, rejected.Body.String())
	}
	var status, reviewNote string
	if err := db.QueryRow("SELECT status, review_note FROM developer_applications WHERE id = ?", submittedBody.Application.ID).Scan(&status, &reviewNote); err != nil {
		t.Fatalf("query rejected application: %v", err)
	}
	if status != "rejected" || reviewNote != "Please add more production references." {
		t.Fatalf("rejected application status=%q note=%q", status, reviewNote)
	}

	rereject := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/admin/developer-applications/"+strconv.FormatInt(submittedBody.Application.ID, 10)+"/reject",
		`{"reviewNote":"Again."}`,
		adminToken,
	)
	if rereject.Code != http.StatusConflict {
		t.Fatalf("rereject status = %d, want %d", rereject.Code, http.StatusConflict)
	}
}

func TestAppPublishReviewAndMarketplaceFlow(t *testing.T) {
	db, app := newAuthTestApp(t)
	userID := insertVerifiedUser(t, db, "dev@example.com", "old-password")
	regularID := insertVerifiedUser(t, db, "buyer@example.com", "old-password")
	adminID := insertVerifiedUser(t, db, "admin@example.com", "old-password")
	if _, err := db.Exec("UPDATE users SET role = 'admin' WHERE id = ?", adminID); err != nil {
		t.Fatalf("mark admin user: %v", err)
	}
	if _, err := db.Exec("INSERT INTO developers(user_id, display_name) VALUES (?, ?)", userID, "Dev Studio"); err != nil {
		t.Fatalf("insert developer: %v", err)
	}

	developerToken := loginToken(t, app, "dev@example.com", "old-password")
	regularToken := loginToken(t, app, "buyer@example.com", "old-password")
	adminToken := loginToken(t, app, "admin@example.com", "old-password")
	body := `{"name":"Build Lens","slug":"build-lens","tagline":"Release intelligence for busy engineering teams.","description":"Track release health, ownership, and launch readiness in one clean workflow.","category":"developer-tools","priceCents":4900,"currency":"USD","iconUrl":"https://example.com/icon.png","coverImageUrl":"https://example.com/cover.png","demoUrl":"https://example.com/demo","docsUrl":"https://example.com/docs","sourceUrl":"https://github.com/example/build-lens","supportUrl":"mailto:support@example.com","tags":["release","analytics"],"version":"1.0.0","releaseNotes":"Initial marketplace release."}`

	unauthenticated := performJSONRequest(t, app, http.MethodPost, "/api/v1/apps", body)
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated publish status = %d, want %d", unauthenticated.Code, http.StatusUnauthorized)
	}
	notDeveloper := performJSONRequestWithToken(t, app, http.MethodPost, "/api/v1/apps", body, regularToken)
	if notDeveloper.Code != http.StatusForbidden {
		t.Fatalf("regular publish status = %d, want %d", notDeveloper.Code, http.StatusForbidden)
	}

	submitted := performJSONRequestWithToken(t, app, http.MethodPost, "/api/v1/apps", body, developerToken)
	if submitted.Code != http.StatusCreated {
		t.Fatalf("publish status = %d, want %d; body = %s", submitted.Code, http.StatusCreated, submitted.Body.String())
	}
	var submittedBody struct {
		App struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
			Slug   string `json:"slug"`
		} `json:"app"`
	}
	if err := json.NewDecoder(submitted.Body).Decode(&submittedBody); err != nil {
		t.Fatalf("decode published app: %v", err)
	}
	if submittedBody.App.ID < 1 || submittedBody.App.Status != "pending_review" || submittedBody.App.Slug != "build-lens" {
		t.Fatalf("submitted app = %+v, want pending_review build-lens", submittedBody.App)
	}

	beforeApproval := performJSONRequest(t, app, http.MethodGet, "/api/v1/marketplace/apps", "")
	if beforeApproval.Code != http.StatusOK {
		t.Fatalf("public list before approval status = %d, want %d", beforeApproval.Code, http.StatusOK)
	}
	var beforeBody struct {
		Apps []struct {
			Slug string `json:"slug"`
		} `json:"apps"`
	}
	if err := json.NewDecoder(beforeApproval.Body).Decode(&beforeBody); err != nil {
		t.Fatalf("decode before approval list: %v", err)
	}
	if len(beforeBody.Apps) != 0 {
		t.Fatalf("public apps before approval = %+v, want empty", beforeBody.Apps)
	}

	nonAdminReview := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/admin/apps/"+strconv.FormatInt(submittedBody.App.ID, 10)+"/approve",
		`{"reviewNote":"Looks good."}`,
		developerToken,
	)
	if nonAdminReview.Code != http.StatusForbidden {
		t.Fatalf("non-admin app approval status = %d, want %d", nonAdminReview.Code, http.StatusForbidden)
	}

	adminList := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/admin/apps", "", adminToken)
	if adminList.Code != http.StatusOK {
		t.Fatalf("admin app list status = %d, want %d; body = %s", adminList.Code, http.StatusOK, adminList.Body.String())
	}
	var adminListBody struct {
		Apps []struct {
			ID             int64  `json:"id"`
			DeveloperEmail string `json:"developerEmail"`
			Status         string `json:"status"`
		} `json:"apps"`
	}
	if err := json.NewDecoder(adminList.Body).Decode(&adminListBody); err != nil {
		t.Fatalf("decode admin app list: %v", err)
	}
	if len(adminListBody.Apps) != 1 || adminListBody.Apps[0].DeveloperEmail != "dev@example.com" || adminListBody.Apps[0].Status != "pending_review" {
		t.Fatalf("admin apps = %+v, want one pending dev@example.com app", adminListBody.Apps)
	}

	approved := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/admin/apps/"+strconv.FormatInt(submittedBody.App.ID, 10)+"/approve",
		`{"reviewNote":"Approved for launch."}`,
		adminToken,
	)
	if approved.Code != http.StatusOK {
		t.Fatalf("approve app status = %d, want %d; body = %s", approved.Code, http.StatusOK, approved.Body.String())
	}
	var status string
	var published bool
	if err := db.QueryRow("SELECT status, published_at IS NOT NULL FROM apps WHERE id = ?", submittedBody.App.ID).Scan(&status, &published); err != nil {
		t.Fatalf("query approved app: %v", err)
	}
	if status != "approved" || !published {
		t.Fatalf("approved app status=%q published=%v, want approved and published", status, published)
	}

	reapprove := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/admin/apps/"+strconv.FormatInt(submittedBody.App.ID, 10)+"/approve",
		`{"reviewNote":"Again."}`,
		adminToken,
	)
	if reapprove.Code != http.StatusConflict {
		t.Fatalf("reapprove app status = %d, want %d", reapprove.Code, http.StatusConflict)
	}

	publicList := performJSONRequest(t, app, http.MethodGet, "/api/v1/marketplace/apps?category=developer-tools&q=build", "")
	if publicList.Code != http.StatusOK {
		t.Fatalf("public list status = %d, want %d; body = %s", publicList.Code, http.StatusOK, publicList.Body.String())
	}
	var publicListBody struct {
		Apps []struct {
			Slug          string   `json:"slug"`
			DeveloperName string   `json:"developerName"`
			Tags          []string `json:"tags"`
		} `json:"apps"`
	}
	if err := json.NewDecoder(publicList.Body).Decode(&publicListBody); err != nil {
		t.Fatalf("decode public list: %v", err)
	}
	if len(publicListBody.Apps) != 1 || publicListBody.Apps[0].Slug != "build-lens" || publicListBody.Apps[0].DeveloperName != "Dev Studio" {
		t.Fatalf("public apps = %+v, want approved build-lens", publicListBody.Apps)
	}

	publicDetail := performJSONRequest(t, app, http.MethodGet, "/api/v1/marketplace/apps/build-lens", "")
	if publicDetail.Code != http.StatusOK {
		t.Fatalf("public detail status = %d, want %d; body = %s", publicDetail.Code, http.StatusOK, publicDetail.Body.String())
	}
	_ = regularID
}

func TestDeveloperAppManagementFlow(t *testing.T) {
	db, app := newAuthTestApp(t)
	userID := insertVerifiedUser(t, db, "dev@example.com", "old-password")
	otherUserID := insertVerifiedUser(t, db, "other@example.com", "old-password")
	regularID := insertVerifiedUser(t, db, "buyer@example.com", "old-password")
	adminID := insertVerifiedUser(t, db, "admin@example.com", "old-password")
	if _, err := db.Exec("UPDATE users SET role = 'admin' WHERE id = ?", adminID); err != nil {
		t.Fatalf("mark admin user: %v", err)
	}
	if _, err := db.Exec("INSERT INTO developers(user_id, display_name) VALUES (?, ?)", userID, "Dev Studio"); err != nil {
		t.Fatalf("insert developer: %v", err)
	}
	if _, err := db.Exec("INSERT INTO developers(user_id, display_name) VALUES (?, ?)", otherUserID, "Other Studio"); err != nil {
		t.Fatalf("insert other developer: %v", err)
	}
	developerToken := loginToken(t, app, "dev@example.com", "old-password")
	regularToken := loginToken(t, app, "buyer@example.com", "old-password")
	adminToken := loginToken(t, app, "admin@example.com", "old-password")
	body := `{"name":"Build Lens","slug":"build-lens","tagline":"Release intelligence for busy engineering teams.","description":"Track release health, ownership, and launch readiness in one clean workflow.","category":"developer-tools","priceCents":4900,"currency":"USD","iconUrl":"https://example.com/icon.png","coverImageUrl":"https://example.com/cover.png","demoUrl":"https://example.com/demo","docsUrl":"https://example.com/docs","sourceUrl":"https://github.com/example/build-lens","supportUrl":"mailto:support@example.com","tags":["release","analytics"],"version":"1.0.0","releaseNotes":"Initial marketplace release."}`
	submitted := performJSONRequestWithToken(t, app, http.MethodPost, "/api/v1/apps", body, developerToken)
	if submitted.Code != http.StatusCreated {
		t.Fatalf("publish status = %d, want %d; body = %s", submitted.Code, http.StatusCreated, submitted.Body.String())
	}
	var submittedBody struct {
		App struct {
			ID int64 `json:"id"`
		} `json:"app"`
	}
	if err := json.NewDecoder(submitted.Body).Decode(&submittedBody); err != nil {
		t.Fatalf("decode published app: %v", err)
	}
	otherResult, err := db.Exec(
		`INSERT INTO apps(developer_id, name, slug, tagline, description, category, price_cents, currency, version)
		 VALUES ((SELECT id FROM developers WHERE user_id = ?), 'Other App', 'other-app', 'Other tagline.', 'Other description', 'saas', 1000, 'USD', '1.0.0')`,
		otherUserID,
	)
	if err != nil {
		t.Fatalf("insert other app: %v", err)
	}
	otherAppID, err := otherResult.LastInsertId()
	if err != nil {
		t.Fatalf("read other app id: %v", err)
	}

	unauthenticated := performJSONRequest(t, app, http.MethodGet, "/api/v1/developer/apps", "")
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated list status = %d, want %d", unauthenticated.Code, http.StatusUnauthorized)
	}
	notDeveloper := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/developer/apps", "", regularToken)
	if notDeveloper.Code != http.StatusForbidden {
		t.Fatalf("regular list status = %d, want %d", notDeveloper.Code, http.StatusForbidden)
	}
	ownedList := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/developer/apps", "", developerToken)
	if ownedList.Code != http.StatusOK {
		t.Fatalf("developer list status = %d, want %d; body = %s", ownedList.Code, http.StatusOK, ownedList.Body.String())
	}
	var ownedListBody struct {
		Apps []struct {
			ID   int64  `json:"id"`
			Slug string `json:"slug"`
		} `json:"apps"`
	}
	if err := json.NewDecoder(ownedList.Body).Decode(&ownedListBody); err != nil {
		t.Fatalf("decode developer apps: %v", err)
	}
	if len(ownedListBody.Apps) != 1 || ownedListBody.Apps[0].Slug != "build-lens" {
		t.Fatalf("developer apps = %+v, want only owned build-lens", ownedListBody.Apps)
	}
	otherDetail := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/developer/apps/"+strconv.FormatInt(otherAppID, 10), "", developerToken)
	if otherDetail.Code != http.StatusNotFound {
		t.Fatalf("other developer app detail status = %d, want %d", otherDetail.Code, http.StatusNotFound)
	}

	approved := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/admin/apps/"+strconv.FormatInt(submittedBody.App.ID, 10)+"/approve",
		`{"reviewNote":"Approved for launch."}`,
		adminToken,
	)
	if approved.Code != http.StatusOK {
		t.Fatalf("approve app status = %d, want %d; body = %s", approved.Code, http.StatusOK, approved.Body.String())
	}
	updateBody := `{"name":"Build Lens Pro","slug":"build-lens-pro","tagline":"Sharper release intelligence.","description":"Updated launch readiness, ownership, and release health details.","category":"developer-tools","priceCents":9900,"currency":"USD","iconUrl":"https://example.com/new-icon.png","coverImageUrl":"https://example.com/new-cover.png","demoUrl":"https://example.com/new-demo","docsUrl":"https://example.com/new-docs","sourceUrl":"https://github.com/example/build-lens-pro","supportUrl":"mailto:help@example.com","tags":["release","governance"],"version":"1.1.0","releaseNotes":"Updated positioning and assets."}`
	updated := performJSONRequestWithToken(t, app, http.MethodPut, "/api/v1/developer/apps/"+strconv.FormatInt(submittedBody.App.ID, 10), updateBody, developerToken)
	if updated.Code != http.StatusOK {
		t.Fatalf("update app status = %d, want %d; body = %s", updated.Code, http.StatusOK, updated.Body.String())
	}
	var updatedBody struct {
		App struct {
			Status        string `json:"status"`
			Slug          string `json:"slug"`
			PriceCents    int64  `json:"priceCents"`
			IconURL       string `json:"iconUrl"`
			CoverImageURL string `json:"coverImageUrl"`
		} `json:"app"`
	}
	if err := json.NewDecoder(updated.Body).Decode(&updatedBody); err != nil {
		t.Fatalf("decode updated app: %v", err)
	}
	if updatedBody.App.Status != "pending_review" || updatedBody.App.Slug != "build-lens-pro" || updatedBody.App.PriceCents != 9900 ||
		updatedBody.App.IconURL != "https://example.com/new-icon.png" || updatedBody.App.CoverImageURL != "https://example.com/new-cover.png" {
		t.Fatalf("updated app = %+v, want pending_review edited assets and price", updatedBody.App)
	}
	publicOldDetail := performJSONRequest(t, app, http.MethodGet, "/api/v1/marketplace/apps/build-lens", "")
	if publicOldDetail.Code != http.StatusNotFound {
		t.Fatalf("old public detail after edit status = %d, want %d", publicOldDetail.Code, http.StatusNotFound)
	}
	publicNewDetail := performJSONRequest(t, app, http.MethodGet, "/api/v1/marketplace/apps/build-lens-pro", "")
	if publicNewDetail.Code != http.StatusNotFound {
		t.Fatalf("new public detail before reapproval status = %d, want %d", publicNewDetail.Code, http.StatusNotFound)
	}
	unauthorizedUpdate := performJSONRequestWithToken(t, app, http.MethodPut, "/api/v1/developer/apps/"+strconv.FormatInt(otherAppID, 10), updateBody, developerToken)
	if unauthorizedUpdate.Code != http.StatusNotFound {
		t.Fatalf("other developer update status = %d, want %d", unauthorizedUpdate.Code, http.StatusNotFound)
	}

	reapproved := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/admin/apps/"+strconv.FormatInt(submittedBody.App.ID, 10)+"/approve",
		`{"reviewNote":"Approved again."}`,
		adminToken,
	)
	if reapproved.Code != http.StatusOK {
		t.Fatalf("reapprove app status = %d, want %d; body = %s", reapproved.Code, http.StatusOK, reapproved.Body.String())
	}
	delisted := performJSONRequestWithToken(t, app, http.MethodPost, "/api/v1/developer/apps/"+strconv.FormatInt(submittedBody.App.ID, 10)+"/delist", `{}`, developerToken)
	if delisted.Code != http.StatusOK {
		t.Fatalf("delist app status = %d, want %d; body = %s", delisted.Code, http.StatusOK, delisted.Body.String())
	}
	var delistedBody struct {
		App struct {
			Status string `json:"status"`
		} `json:"app"`
	}
	if err := json.NewDecoder(delisted.Body).Decode(&delistedBody); err != nil {
		t.Fatalf("decode delisted app: %v", err)
	}
	if delistedBody.App.Status != "delisted" {
		t.Fatalf("delisted status = %q, want delisted", delistedBody.App.Status)
	}
	repeatedDelist := performJSONRequestWithToken(t, app, http.MethodPost, "/api/v1/developer/apps/"+strconv.FormatInt(submittedBody.App.ID, 10)+"/delist", `{}`, developerToken)
	if repeatedDelist.Code != http.StatusOK {
		t.Fatalf("repeated delist app status = %d, want %d", repeatedDelist.Code, http.StatusOK)
	}
	publicList := performJSONRequest(t, app, http.MethodGet, "/api/v1/marketplace/apps?q=build", "")
	if publicList.Code != http.StatusOK {
		t.Fatalf("public list status = %d, want %d; body = %s", publicList.Code, http.StatusOK, publicList.Body.String())
	}
	var publicListBody struct {
		Apps []struct {
			Slug string `json:"slug"`
		} `json:"apps"`
	}
	if err := json.NewDecoder(publicList.Body).Decode(&publicListBody); err != nil {
		t.Fatalf("decode public apps: %v", err)
	}
	if len(publicListBody.Apps) != 0 {
		t.Fatalf("public apps after delist = %+v, want empty", publicListBody.Apps)
	}
	publicDelistedDetail := performJSONRequest(t, app, http.MethodGet, "/api/v1/marketplace/apps/build-lens-pro", "")
	if publicDelistedDetail.Code != http.StatusNotFound {
		t.Fatalf("public detail after delist status = %d, want %d", publicDelistedDetail.Code, http.StatusNotFound)
	}
	_ = regularID
}

func TestAppRejectionRequiresNote(t *testing.T) {
	db, app := newAuthTestApp(t)
	userID := insertVerifiedUser(t, db, "dev@example.com", "old-password")
	adminID := insertVerifiedUser(t, db, "admin@example.com", "old-password")
	if _, err := db.Exec("UPDATE users SET role = 'admin' WHERE id = ?", adminID); err != nil {
		t.Fatalf("mark admin user: %v", err)
	}
	if _, err := db.Exec("INSERT INTO developers(user_id, display_name) VALUES (?, ?)", userID, "Dev Studio"); err != nil {
		t.Fatalf("insert developer: %v", err)
	}
	adminToken := loginToken(t, app, "admin@example.com", "old-password")
	result, err := db.Exec(
		`INSERT INTO apps(developer_id, name, slug, tagline, description, category, price_cents, currency, version)
		 VALUES ((SELECT id FROM developers WHERE user_id = ?), 'Draft App', 'draft-app', 'A useful draft.', 'Draft description', 'saas', 0, 'USD', '0.1.0')`,
		userID,
	)
	if err != nil {
		t.Fatalf("insert draft app: %v", err)
	}
	appID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read draft app id: %v", err)
	}

	missingNote := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/admin/apps/"+strconv.FormatInt(appID, 10)+"/reject",
		`{"reviewNote":""}`,
		adminToken,
	)
	if missingNote.Code != http.StatusBadRequest {
		t.Fatalf("missing note rejection status = %d, want %d", missingNote.Code, http.StatusBadRequest)
	}
	rejected := performJSONRequestWithToken(
		t,
		app,
		http.MethodPost,
		"/api/v1/admin/apps/"+strconv.FormatInt(appID, 10)+"/reject",
		`{"reviewNote":"Needs clearer docs."}`,
		adminToken,
	)
	if rejected.Code != http.StatusOK {
		t.Fatalf("reject status = %d, want %d; body = %s", rejected.Code, http.StatusOK, rejected.Body.String())
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
	cfg.AvatarUploadDir = t.TempDir()
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

func loginToken(t *testing.T, app http.Handler, email string, password string) string {
	t.Helper()
	login := performJSONRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/login",
		`{"email":"`+email+`","password":"`+password+`"}`,
	)
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d; body = %s", login.Code, http.StatusOK, login.Body.String())
	}
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(login.Body).Decode(&body); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	return body.Token
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

func performMultipartAvatarRequest(
	t *testing.T,
	app http.Handler,
	token string,
	avatar []byte,
) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("avatar", "avatar.png")
	if err != nil {
		t.Fatalf("create avatar form file: %v", err)
	}
	if _, err := part.Write(avatar); err != nil {
		t.Fatalf("write avatar form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPatch, "/api/v1/auth/me/avatar", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	return response
}

func validPNGAvatar() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89,
	}
}
