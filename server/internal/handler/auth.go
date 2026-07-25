package handler

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"devkit/server/internal/config"
	"devkit/server/internal/middleware"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	maxRequestBodySize    = 1 << 20
	maxAvatarUploadSize   = 2 << 20
	verificationLifetime  = 24 * time.Hour
	passwordResetLifetime = 1 * time.Hour
)

type Auth struct {
	db     *sql.DB
	config config.AuthConfig
	logger *slog.Logger
	now    func() time.Time
}

type authUser struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	AvatarURL   string `json:"avatarUrl"`
}

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

type registerResponse struct {
	Message string   `json:"message"`
	User    authUser `json:"user"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

type changePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type loginResponse struct {
	Token     string   `json:"token"`
	ExpiresAt string   `json:"expiresAt"`
	User      authUser `json:"user"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type authClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func NewAuth(db *sql.DB, cfg config.AuthConfig, logger *slog.Logger) *Auth {
	return &Auth{
		db:     db,
		config: cfg,
		logger: logger,
		now:    time.Now,
	}
}

func (h *Auth) Register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.DisplayName = strings.TrimSpace(request.DisplayName)
	if err := validateRegistration(request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "hash registration password", "error", err)
		writeError(w, http.StatusInternalServerError, "could not register user")
		return
	}

	avatarURL := defaultAvatarURL(request.Email)
	verificationToken := uuid.NewString()
	expiresAt := h.now().UTC().Add(verificationLifetime)

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "begin registration", "error", err)
		writeError(w, http.StatusInternalServerError, "could not register user")
		return
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(
		r.Context(),
		`INSERT INTO users(email, password_hash, avatar_url, display_name)
		 VALUES (?, ?, ?, ?)`,
		request.Email,
		string(passwordHash),
		avatarURL,
		request.DisplayName,
	)
	if err != nil {
		if isUniqueEmailError(err) {
			writeError(w, http.StatusConflict, "email is already registered")
			return
		}
		h.logger.ErrorContext(r.Context(), "insert registered user", "error", err)
		writeError(w, http.StatusInternalServerError, "could not register user")
		return
	}

	userID, err := result.LastInsertId()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read registered user ID", "error", err)
		writeError(w, http.StatusInternalServerError, "could not register user")
		return
	}
	if _, err := tx.ExecContext(
		r.Context(),
		`INSERT INTO user_verifications(token, user_id, expires_at) VALUES (?, ?, ?)`,
		verificationToken,
		userID,
		expiresAt,
	); err != nil {
		h.logger.ErrorContext(r.Context(), "insert verification token", "error", err)
		writeError(w, http.StatusInternalServerError, "could not register user")
		return
	}
	if err := tx.Commit(); err != nil {
		h.logger.ErrorContext(r.Context(), "commit registration", "error", err)
		writeError(w, http.StatusInternalServerError, "could not register user")
		return
	}

	activationLink := strings.TrimRight(h.config.ActivationBaseURL, "/") +
		"/api/v1/auth/activate?token=" + url.QueryEscape(verificationToken)
	if err := h.sendActivationEmail(request.Email, request.DisplayName, activationLink); err != nil {
		h.logger.ErrorContext(r.Context(), "send activation email", "email", request.Email, "error", err)
		writeError(w, http.StatusInternalServerError, "registration created but activation email could not be sent")
		return
	}

	writeJSON(w, http.StatusCreated, registerResponse{
		Message: "registration successful; check your email for the activation link",
		User: authUser{
			ID:          userID,
			Email:       request.Email,
			DisplayName: request.DisplayName,
			AvatarURL:   avatarURL,
		},
	})
}

func (h *Auth) sendActivationEmail(to string, displayName string, activationLink string) error {
	smtpConfig := h.config.SMTP
	if smtpConfig.Host == "" || smtpConfig.From == "" {
		log.Printf("Activation link: %s", activationLink)
		return nil
	}
	if !validEmail(smtpConfig.From) {
		return fmt.Errorf("invalid SMTP from address")
	}

	subject := "Activate your DevKit account"
	body := fmt.Sprintf(
		"Hi %s,\r\n\r\nActivate your DevKit account with this link:\r\n%s\r\n\r\nThis link expires in 24 hours.\r\n",
		displayName,
		activationLink,
	)
	message := strings.Join([]string{
		"From: " + smtpConfig.From,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	address := net.JoinHostPort(smtpConfig.Host, strconv.Itoa(smtpConfig.Port))
	var auth smtp.Auth
	if smtpConfig.Username != "" || smtpConfig.Password != "" {
		auth = smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Host)
	}
	return smtp.SendMail(address, auth, smtpConfig.From, []string{to}, []byte(message))
}

func (h *Auth) sendPasswordResetEmail(to string, displayName string, resetLink string) error {
	smtpConfig := h.config.SMTP
	if smtpConfig.Host == "" || smtpConfig.From == "" {
		log.Printf("Password reset link: %s", resetLink)
		return nil
	}
	if !validEmail(smtpConfig.From) {
		return fmt.Errorf("invalid SMTP from address")
	}

	subject := "Reset your DevKit password"
	body := fmt.Sprintf(
		"Hi %s,\r\n\r\nReset your DevKit password with this link:\r\n%s\r\n\r\nThis link expires in 1 hour. If you did not request this, you can ignore this email.\r\n",
		displayName,
		resetLink,
	)
	message := strings.Join([]string{
		"From: " + smtpConfig.From,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	address := net.JoinHostPort(smtpConfig.Host, strconv.Itoa(smtpConfig.Port))
	var auth smtp.Auth
	if smtpConfig.Username != "" || smtpConfig.Password != "" {
		auth = smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Host)
	}
	return smtp.SendMail(address, auth, smtpConfig.From, []string{to}, []byte(message))
}

func (h *Auth) Activate(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if _, err := uuid.Parse(token); err != nil {
		writeError(w, http.StatusBadRequest, "invalid activation token")
		return
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "begin account activation", "error", err)
		writeError(w, http.StatusInternalServerError, "could not activate user")
		return
	}
	defer tx.Rollback()

	var userID int64
	var expiresAt time.Time
	err = tx.QueryRowContext(
		r.Context(),
		"SELECT user_id, expires_at FROM user_verifications WHERE token = ?",
		token,
	).Scan(&userID, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusBadRequest, "invalid or expired activation token")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read activation token", "error", err)
		writeError(w, http.StatusInternalServerError, "could not activate user")
		return
	}
	if !expiresAt.After(h.now().UTC()) {
		if _, err := tx.ExecContext(r.Context(), "DELETE FROM user_verifications WHERE token = ?", token); err != nil {
			h.logger.ErrorContext(r.Context(), "delete expired activation token", "error", err)
		}
		if err := tx.Commit(); err != nil {
			h.logger.ErrorContext(r.Context(), "commit expired token deletion", "error", err)
		}
		writeError(w, http.StatusBadRequest, "invalid or expired activation token")
		return
	}

	if _, err := tx.ExecContext(
		r.Context(),
		"UPDATE users SET verified_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		userID,
	); err != nil {
		h.logger.ErrorContext(r.Context(), "verify activated user", "error", err)
		writeError(w, http.StatusInternalServerError, "could not activate user")
		return
	}
	if _, err := tx.ExecContext(
		r.Context(),
		"DELETE FROM user_verifications WHERE token = ?",
		token,
	); err != nil {
		h.logger.ErrorContext(r.Context(), "consume activation token", "error", err)
		writeError(w, http.StatusInternalServerError, "could not activate user")
		return
	}
	if err := tx.Commit(); err != nil {
		h.logger.ErrorContext(r.Context(), "commit account activation", "error", err)
		writeError(w, http.StatusInternalServerError, "could not activate user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "activated"})
}

func (h *Auth) UpgradeToDeveloper(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var verifiedAt sql.NullTime
	if err := h.db.QueryRowContext(
		r.Context(),
		"SELECT verified_at FROM users WHERE id = ?",
		userID,
	).Scan(&verifiedAt); err != nil {
		writeError(w, http.StatusInternalServerError, "could not verify user")
		return
	}
	if !verifiedAt.Valid {
		writeError(w, http.StatusForbidden, "email address is not verified")
		return
	}

	var displayName string
	if err := h.db.QueryRowContext(
		r.Context(),
		"SELECT display_name FROM users WHERE id = ?",
		userID,
	).Scan(&displayName); err != nil {
		writeError(w, http.StatusInternalServerError, "could not read user")
		return
	}

	_, err := h.db.ExecContext(
		r.Context(),
		"INSERT INTO developers(user_id, display_name) VALUES (?, ?)",
		userID,
		displayName,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: developers.user_id") {
			writeError(w, http.StatusConflict, "already a developer")
			return
		}
		h.logger.ErrorContext(r.Context(), "upgrade to developer", "error", err)
		writeError(w, http.StatusInternalServerError, "could not upgrade to developer")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "upgraded"})
}

func (h *Auth) Me(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	user, err := h.readAuthUser(w, r, userID)
	if err != nil {
		return
	}

	writeJSON(w, http.StatusOK, map[string]authUser{"user": user})
}

func (h *Auth) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarUploadSize+1024)
	file, header, err := r.FormFile("avatar")
	if err != nil {
		writeError(w, http.StatusBadRequest, "avatar image file is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxAvatarUploadSize+1))
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read avatar upload", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update avatar")
		return
	}
	if len(data) == 0 {
		writeError(w, http.StatusBadRequest, "avatar image file is required")
		return
	}
	if len(data) > maxAvatarUploadSize {
		writeError(w, http.StatusBadRequest, "avatar image must be 2MB or smaller")
		return
	}

	extension, ok := avatarFileExtension(http.DetectContentType(data))
	if !ok {
		writeError(w, http.StatusBadRequest, "avatar must be a JPEG, PNG, GIF, or WebP image")
		return
	}

	if err := os.MkdirAll(h.config.AvatarUploadDir, 0755); err != nil {
		h.logger.ErrorContext(r.Context(), "create avatar upload directory", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update avatar")
		return
	}

	fileName := fmt.Sprintf("user-%d-%s%s", userID, uuid.NewString(), extension)
	filePath := filepath.Join(h.config.AvatarUploadDir, fileName)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		h.logger.ErrorContext(r.Context(), "write avatar upload", "name", header.Filename, "error", err)
		writeError(w, http.StatusInternalServerError, "could not update avatar")
		return
	}

	avatarURL := "/api/v1/uploads/avatars/" + fileName
	if _, err := h.db.ExecContext(
		r.Context(),
		"UPDATE users SET avatar_url = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		avatarURL,
		userID,
	); err != nil {
		h.logger.ErrorContext(r.Context(), "update avatar URL", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update avatar")
		return
	}

	user, err := h.readAuthUser(w, r, userID)
	if err != nil {
		return
	}

	writeJSON(w, http.StatusOK, map[string]authUser{"user": user})
}

func (h *Auth) ResetAvatar(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var email string
	if err := h.db.QueryRowContext(
		r.Context(),
		"SELECT email FROM users WHERE id = ?",
		userID,
	).Scan(&email); err != nil {
		h.logger.ErrorContext(r.Context(), "read user email for avatar reset", "error", err)
		writeError(w, http.StatusInternalServerError, "could not reset avatar")
		return
	}

	if _, err := h.db.ExecContext(
		r.Context(),
		"UPDATE users SET avatar_url = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		defaultAvatarURL(email),
		userID,
	); err != nil {
		h.logger.ErrorContext(r.Context(), "reset avatar URL", "error", err)
		writeError(w, http.StatusInternalServerError, "could not reset avatar")
		return
	}

	user, err := h.readAuthUser(w, r, userID)
	if err != nil {
		return
	}

	writeJSON(w, http.StatusOK, map[string]authUser{"user": user})
}

func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	if !validEmail(request.Email) || request.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	var user authUser
	var passwordHash string
	var verifiedAt sql.NullTime
	err := h.db.QueryRowContext(
		r.Context(),
		`SELECT id, email, password_hash, display_name, avatar_url, verified_at
		 FROM users WHERE email = ?`,
		request.Email,
	).Scan(
		&user.ID,
		&user.Email,
		&passwordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&verifiedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read login user", "error", err)
		writeError(w, http.StatusInternalServerError, "could not log in")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(request.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if !verifiedAt.Valid {
		writeError(w, http.StatusForbidden, "email address is not verified")
		return
	}

	issuedAt := h.now().UTC()
	expiresAt := issuedAt.Add(time.Duration(h.config.JWTExpiryHours) * time.Hour)
	claims := authClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "devkit",
			Subject:   strconv.FormatInt(user.ID, 10),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(h.config.JWTSecret))
	if err != nil {
		h.logger.ErrorContext(r.Context(), "sign login token", "error", err)
		writeError(w, http.StatusInternalServerError, "could not log in")
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		User:      user,
	})
}

func (h *Auth) readAuthUser(w http.ResponseWriter, r *http.Request, userID int64) (authUser, error) {
	var user authUser
	if err := h.db.QueryRowContext(
		r.Context(),
		`SELECT id, email, display_name, avatar_url
		 FROM users WHERE id = ?`,
		userID,
	).Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
	); err != nil {
		h.logger.ErrorContext(r.Context(), "read auth user", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read user")
		return authUser{}, err
	}
	return user, nil
}

func (h *Auth) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var request forgotPasswordRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	if !validEmail(request.Email) {
		writeError(w, http.StatusBadRequest, "valid email is required")
		return
	}

	response := map[string]string{"message": "if the email is registered, a reset link has been sent"}
	var user struct {
		ID          int64
		DisplayName string
	}
	err := h.db.QueryRowContext(
		r.Context(),
		"SELECT id, display_name FROM users WHERE email = ?",
		request.Email,
	).Scan(&user.ID, &user.DisplayName)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusOK, response)
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read password reset user", "error", err)
		writeError(w, http.StatusInternalServerError, "could not request password reset")
		return
	}

	token := uuid.NewString()
	expiresAt := h.now().UTC().Add(passwordResetLifetime)
	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "begin password reset request", "error", err)
		writeError(w, http.StatusInternalServerError, "could not request password reset")
		return
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(r.Context(), "DELETE FROM password_resets WHERE user_id = ?", user.ID); err != nil {
		h.logger.ErrorContext(r.Context(), "delete existing password resets", "error", err)
		writeError(w, http.StatusInternalServerError, "could not request password reset")
		return
	}
	if _, err := tx.ExecContext(
		r.Context(),
		`INSERT INTO password_resets(token, user_id, expires_at) VALUES (?, ?, ?)`,
		token,
		user.ID,
		expiresAt,
	); err != nil {
		h.logger.ErrorContext(r.Context(), "insert password reset token", "error", err)
		writeError(w, http.StatusInternalServerError, "could not request password reset")
		return
	}
	if err := tx.Commit(); err != nil {
		h.logger.ErrorContext(r.Context(), "commit password reset request", "error", err)
		writeError(w, http.StatusInternalServerError, "could not request password reset")
		return
	}

	resetLink := strings.TrimRight(h.config.ActivationBaseURL, "/") +
		"/reset-password?token=" + url.QueryEscape(token)
	if err := h.sendPasswordResetEmail(request.Email, user.DisplayName, resetLink); err != nil {
		h.logger.ErrorContext(r.Context(), "send password reset email", "email", request.Email, "error", err)
		writeError(w, http.StatusInternalServerError, "could not send password reset email")
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Auth) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var request resetPasswordRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	request.Token = strings.TrimSpace(request.Token)
	if _, err := uuid.Parse(request.Token); err != nil {
		writeError(w, http.StatusBadRequest, "invalid or expired reset token")
		return
	}
	if err := validatePassword(request.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "hash reset password", "error", err)
		writeError(w, http.StatusInternalServerError, "could not reset password")
		return
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "begin password reset", "error", err)
		writeError(w, http.StatusInternalServerError, "could not reset password")
		return
	}
	defer tx.Rollback()

	var userID int64
	var expiresAt time.Time
	err = tx.QueryRowContext(
		r.Context(),
		"SELECT user_id, expires_at FROM password_resets WHERE token = ?",
		request.Token,
	).Scan(&userID, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusBadRequest, "invalid or expired reset token")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "read password reset token", "error", err)
		writeError(w, http.StatusInternalServerError, "could not reset password")
		return
	}
	if !expiresAt.After(h.now().UTC()) {
		if _, err := tx.ExecContext(r.Context(), "DELETE FROM password_resets WHERE token = ?", request.Token); err != nil {
			h.logger.ErrorContext(r.Context(), "delete expired password reset token", "error", err)
		}
		if err := tx.Commit(); err != nil {
			h.logger.ErrorContext(r.Context(), "commit expired password reset deletion", "error", err)
		}
		writeError(w, http.StatusBadRequest, "invalid or expired reset token")
		return
	}

	if _, err := tx.ExecContext(
		r.Context(),
		"UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		string(passwordHash),
		userID,
	); err != nil {
		h.logger.ErrorContext(r.Context(), "update reset password", "error", err)
		writeError(w, http.StatusInternalServerError, "could not reset password")
		return
	}
	if _, err := tx.ExecContext(r.Context(), "DELETE FROM password_resets WHERE token = ?", request.Token); err != nil {
		h.logger.ErrorContext(r.Context(), "consume password reset token", "error", err)
		writeError(w, http.StatusInternalServerError, "could not reset password")
		return
	}
	if err := tx.Commit(); err != nil {
		h.logger.ErrorContext(r.Context(), "commit password reset", "error", err)
		writeError(w, http.StatusInternalServerError, "could not reset password")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "password reset"})
}

func (h *Auth) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var request changePasswordRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if request.OldPassword == "" {
		writeError(w, http.StatusBadRequest, "old password is required")
		return
	}
	if err := validatePassword(request.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var passwordHash string
	if err := h.db.QueryRowContext(
		r.Context(),
		"SELECT password_hash FROM users WHERE id = ?",
		userID,
	).Scan(&passwordHash); err != nil {
		h.logger.ErrorContext(r.Context(), "read password for change", "error", err)
		writeError(w, http.StatusInternalServerError, "could not change password")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(request.OldPassword)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid current password")
		return
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "hash changed password", "error", err)
		writeError(w, http.StatusInternalServerError, "could not change password")
		return
	}
	if _, err := h.db.ExecContext(
		r.Context(),
		"UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		string(newPasswordHash),
		userID,
	); err != nil {
		h.logger.ErrorContext(r.Context(), "update changed password", "error", err)
		writeError(w, http.StatusInternalServerError, "could not change password")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "password changed"})
}

func validateRegistration(request registerRequest) error {
	if !validEmail(request.Email) {
		return errors.New("invalid email address")
	}
	if err := validatePassword(request.Password); err != nil {
		return err
	}
	displayNameLength := utf8.RuneCountInString(request.DisplayName)
	if displayNameLength < 1 || displayNameLength > 80 {
		return errors.New("displayName must be between 1 and 80 characters")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if len(password) > 72 {
		return errors.New("password must not exceed 72 bytes")
	}
	return nil
}

func validEmail(email string) bool {
	if email == "" || len(email) > 254 || strings.ContainsAny(email, "\r\n") {
		return false
	}
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email && strings.Contains(email, "@")
}

func avatarFileExtension(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/gif":
		return ".gif", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}

func defaultAvatarURL(email string) string {
	digest := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(email))))
	return "https://www.gravatar.com/avatar/" + hex.EncodeToString(digest[:]) +
		"?d=identicon&s=240"
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return errors.New("invalid JSON request body")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}

func isUniqueEmailError(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed: users.email")
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func writeJSON(w http.ResponseWriter, status int, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// The response headers have already been sent, so only logging is possible.
		log.Printf("encode JSON response: %v", fmt.Errorf("encode response: %w", err))
	}
}
