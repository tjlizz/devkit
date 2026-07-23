package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type authContextKey string

const (
	userIDContextKey authContextKey = "user_id"
	emailContextKey  authContextKey = "email"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func JWT(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			scheme, tokenText, found := strings.Cut(header, " ")
			if !found || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(tokenText) == "" {
				writeUnauthorized(w, "missing or invalid authorization header")
				return
			}

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(
				strings.TrimSpace(tokenText),
				claims,
				func(token *jwt.Token) (any, error) {
					return []byte(secret), nil
				},
				jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
				jwt.WithIssuer("devkit"),
			)
			if err != nil || !token.Valid || claims.UserID <= 0 || claims.Email == "" {
				writeUnauthorized(w, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
			ctx = context.WithValue(ctx, emailContextKey, claims.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int64)
	return userID, ok
}

func EmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(emailContextKey).(string)
	return email, ok
}

func FromContext(ctx context.Context) (userID int64, email string, ok bool) {
	userID, userOK := UserIDFromContext(ctx)
	email, emailOK := EmailFromContext(ctx)
	return userID, email, userOK && emailOK
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		return
	}
}
