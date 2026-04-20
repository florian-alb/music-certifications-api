package model

import (
	"context"
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type APIKey struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	KeyHash   string     `json:"-"`
	Tier      string     `json:"tier"`
	ReqCount  int64      `json:"req_count"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type ctxKey string

const ctxAPIKeyVal ctxKey = "apikey"

func SetAPIKeyCtx(ctx context.Context, key *APIKey) context.Context {
	return context.WithValue(ctx, ctxAPIKeyVal, key)
}

func APIKeyFromCtx(ctx context.Context) *APIKey {
	v, _ := ctx.Value(ctxAPIKeyVal).(*APIKey)
	return v
}
