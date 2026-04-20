package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/florian-alb/music-certifications/internal/model"
	"github.com/florian-alb/music-certifications/pkg/response"
)

type limiter interface {
	CheckLimit(ctx context.Context, keyID, tier string) (bool, int64, error)
}

type RateLimitMiddleware struct {
	cache limiter
}

func NewRateLimitMiddleware(cache limiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{cache: cache}
}

func (m *RateLimitMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := model.APIKeyFromCtx(r.Context())
		if apiKey == nil {
			// Auth middleware must run before this one
			response.Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		allowed, remaining, err := m.cache.CheckLimit(r.Context(), apiKey.ID, apiKey.Tier)
		if err != nil {
			// Fail open: Redis outage shouldn't take down the API
			next.ServeHTTP(w, r)
			return
		}

		if remaining >= 0 {
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		}

		if !allowed {
			response.Error(w, http.StatusTooManyRequests, "daily rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r)
	})
}
