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
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"devkit/server/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	maxRequestBodySize   = 1 << 20
	verificationLifetime = 24 * time.Hour
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
		"INSERT INTO developers(user_id, display_name) VALUES (?, ?)",
		userID,
		request.DisplayName,
	); err != nil {
		h.logger.ErrorContext(r.Context(), "insert registered developer", "error", err)
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
	log.Printf("Activation link: %s", activationLink)

	writeJSON(w, http.StatusCreated, registerResponse{
		Message: "registration successful; check the server log for the activation link",
		User: authUser{
			ID:          userID,
			Email:       request.Email,
			DisplayName: request.DisplayName,
			AvatarURL:   avatarURL,
		},
	})
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

func validateRegistration(request registerRequest) error {
	if !validEmail(request.Email) {
		return errors.New("invalid email address")
	}
	if len(request.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if len(request.Password) > 72 {
		return errors.New("password must not exceed 72 bytes")
	}
	displayNameLength := utf8.RuneCountInString(request.DisplayName)
	if displayNameLength < 1 || displayNameLength > 80 {
		return errors.New("displayName must be between 1 and 80 characters")
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
