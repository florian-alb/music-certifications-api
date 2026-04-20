package middleware

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/florian-alb/music-certifications/internal/model"
	"github.com/florian-alb/music-certifications/pkg/response"
)

type keyLookup interface {
	GetKeyByHash(ctx context.Context, plainKey string) (*model.APIKey, error)
}

type AuthMiddleware struct {
	auth keyLookup
}

func NewAuthMiddleware(auth keyLookup) *AuthMiddleware {
	return &AuthMiddleware{auth: auth}
}

func (m *AuthMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		plainKey := r.Header.Get("X-API-Key")
		if plainKey == "" {
			response.Error(w, http.StatusUnauthorized, "missing X-API-Key header")
			return
		}

		apiKey, err := m.auth.GetKeyByHash(r.Context(), plainKey)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				response.Error(w, http.StatusUnauthorized, "invalid API key")
				return
			}
			response.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}

		if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
			response.Error(w, http.StatusUnauthorized, "API key expired")
			return
		}

		ctx := model.SetAPIKeyCtx(r.Context(), apiKey)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
