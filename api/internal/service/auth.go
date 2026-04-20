package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/florian-alb/music-certifications/internal/model"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken   = errors.New("email already in use")
	ErrInvalidCreds = errors.New("invalid credentials")
)

type userRepo interface {
	Create(ctx context.Context, email, passwordHash string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

type apiKeyRepo interface {
	Create(ctx context.Context, userID, keyHash, tier string) (*model.APIKey, error)
	GetByHash(ctx context.Context, keyHash string) (*model.APIKey, error)
	GetByUserID(ctx context.Context, userID string) (*model.APIKey, error)
	DeleteByUserID(ctx context.Context, userID string) error
}

type AuthService struct {
	users   userRepo
	apiKeys apiKeyRepo
}

func NewAuthService(users userRepo, apiKeys apiKeyRepo) *AuthService {
	return &AuthService{users: users, apiKeys: apiKeys}
}

type AuthResult struct {
	User     *model.User   `json:"user"`
	APIKey   *model.APIKey `json:"api_key"`
	PlainKey string        `json:"plain_key"`
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*AuthResult, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.users.Create(ctx, email, string(hash))
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ErrEmailTaken
		}
		return nil, err
	}

	plain, keyHash, err := newAPIKey()
	if err != nil {
		return nil, err
	}

	key, err := s.apiKeys.Create(ctx, user.ID, keyHash, "free")
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: user, APIKey: key, PlainKey: plain}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCreds
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCreds
	}

	// Rotate: delete old key, issue new one
	if err := s.apiKeys.DeleteByUserID(ctx, user.ID); err != nil {
		return nil, err
	}

	plain, keyHash, err := newAPIKey()
	if err != nil {
		return nil, err
	}

	key, err := s.apiKeys.Create(ctx, user.ID, keyHash, "free")
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: user, APIKey: key, PlainKey: plain}, nil
}

func (s *AuthService) GetMyKey(ctx context.Context, userID string) (*model.APIKey, error) {
	return s.apiKeys.GetByUserID(ctx, userID)
}

func (s *AuthService) GetKeyByHash(ctx context.Context, plainKey string) (*model.APIKey, error) {
	return s.apiKeys.GetByHash(ctx, hashKey(plainKey))
}

func newAPIKey() (plain, keyHash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate key: %w", err)
	}
	plain = hex.EncodeToString(b)
	return plain, hashKey(plain), nil
}

func hashKey(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
