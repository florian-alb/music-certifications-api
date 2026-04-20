package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/florian-alb/music-certifications/internal/model"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, email, passwordHash string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email, created_at`,
		email, passwordHash,
	).Scan(&u.ID, &u.Email, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, created_at FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &u, nil
}

type APIKeyRepo struct {
	db *sql.DB
}

func NewAPIKeyRepo(db *sql.DB) *APIKeyRepo {
	return &APIKeyRepo{db: db}
}

func (r *APIKeyRepo) Create(ctx context.Context, userID, keyHash, tier string) (*model.APIKey, error) {
	var k model.APIKey
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO api_keys (user_id, key_hash, tier) VALUES ($1, $2, $3)
		 RETURNING id, user_id, tier, req_count, created_at`,
		userID, keyHash, tier,
	).Scan(&k.ID, &k.UserID, &k.Tier, &k.ReqCount, &k.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create api key: %w", err)
	}
	return &k, nil
}

func (r *APIKeyRepo) GetByHash(ctx context.Context, keyHash string) (*model.APIKey, error) {
	var k model.APIKey
	var expiresAt sql.NullTime
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, key_hash, tier, req_count, created_at, expires_at
		 FROM api_keys WHERE key_hash = $1`, keyHash,
	).Scan(&k.ID, &k.UserID, &k.KeyHash, &k.Tier, &k.ReqCount, &k.CreatedAt, &expiresAt)
	if err != nil {
		return nil, fmt.Errorf("get api key: %w", err)
	}
	if expiresAt.Valid {
		k.ExpiresAt = &expiresAt.Time
	}
	return &k, nil
}

func (r *APIKeyRepo) GetByUserID(ctx context.Context, userID string) (*model.APIKey, error) {
	var k model.APIKey
	var expiresAt sql.NullTime
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, key_hash, tier, req_count, created_at, expires_at
		 FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`, userID,
	).Scan(&k.ID, &k.UserID, &k.KeyHash, &k.Tier, &k.ReqCount, &k.CreatedAt, &expiresAt)
	if err != nil {
		return nil, fmt.Errorf("get api key by user: %w", err)
	}
	if expiresAt.Valid {
		k.ExpiresAt = &expiresAt.Time
	}
	return &k, nil
}

func (r *APIKeyRepo) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM api_keys WHERE user_id = $1`, userID)
	return err
}
