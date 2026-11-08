package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/rucodencode/shortener/internal/auth"
)

const (
	cookieName   = "visitor_id"
	cookieMaxAge = 24 * time.Hour
)

type contextKey struct{}

type AuthMiddleware struct {
	tokenManager auth.TokenManager
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(contextKey{}).(string)
	return userID, ok && userID != ""
}

func NewAuthMiddleware(tokenManager auth.TokenManager) *AuthMiddleware {
	return &AuthMiddleware{
		tokenManager: tokenManager,
	}
}

func (m *AuthMiddleware) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string

		cookie, err := r.Cookie(cookieName)

		if err == nil {
			userID, err = m.tokenManager.Parse(cookie.Value)
			if err != nil {
				userID = ""
			} else if userID == "" {
				http.Error(w, "", http.StatusUnauthorized)
				return
			}
		}

		if userID == "" {
			userID, err = generateUserID()
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			token, err := m.tokenManager.Generate(userID)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Value:    token,
				Path:     "/",
				MaxAge:   int(cookieMaxAge.Seconds()),
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			})
		}

		ctx := context.WithValue(r.Context(), contextKey{}, userID)

		h.ServeHTTP(w, r.WithContext(ctx))

	})
}

func generateUserID() (string, error) {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return "", err
	}

	return hex.EncodeToString(id), nil
}
