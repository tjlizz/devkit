package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTInjectsUserIdentity(t *testing.T) {
	const secret = "middleware-test-secret"
	claims := Claims{
		UserID: 42,
		Email:  "dev@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "devkit",
			Subject:   strconv.FormatInt(42, 10),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign test JWT: %v", err)
	}

	protected := JWT(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, hasUserID := UserIDFromContext(r.Context())
		email, hasEmail := EmailFromContext(r.Context())
		if !hasUserID || !hasEmail {
			t.Error("identity was not injected into context")
		}
		if userID != 42 || email != "dev@example.com" {
			t.Errorf("context identity = %d, %q", userID, email)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()

	protected.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestJWTRejectsInvalidCredentials(t *testing.T) {
	protected := JWT("right-secret")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("protected handler must not be called")
		w.WriteHeader(http.StatusNoContent)
	}))

	tests := []struct {
		name   string
		header string
	}{
		{name: "missing"},
		{name: "wrong scheme", header: "Basic abc"},
		{name: "invalid token", header: "Bearer not-a-token"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			request.Header.Set("Authorization", test.header)
			response := httptest.NewRecorder()
			protected.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want %d", response.Code, http.StatusUnauthorized)
			}
		})
	}
}
