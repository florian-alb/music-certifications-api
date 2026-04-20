package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/florian-alb/music-certifications/internal/model"
	"github.com/florian-alb/music-certifications/internal/service"
	"github.com/florian-alb/music-certifications/pkg/response"
)

type authService interface {
	Register(ctx context.Context, email, password string) (*service.AuthResult, error)
	Login(ctx context.Context, email, password string) (*service.AuthResult, error)
	GetMyKey(ctx context.Context, userID string) (*model.APIKey, error)
}

type AuthHandler struct {
	svc authService
}

func NewAuthHandler(svc authService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Credentials is the request body for auth endpoints.
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body Credentials true "Email and password"
// @Success      201  {object}  service.AuthResult
// @Failure      400  {object}  response.ErrorResponse
// @Failure      409  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /v1/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if creds.Email == "" || creds.Password == "" {
		response.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	result, err := h.svc.Register(r.Context(), creds.Email, creds.Password)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			response.Error(w, http.StatusConflict, "email already in use")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	response.JSON(w, http.StatusCreated, result)
}

// Login godoc
// @Summary      Login and rotate API key
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body Credentials true "Email and password"
// @Success      200  {object}  service.AuthResult
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.Login(r.Context(), creds.Email, creds.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCreds) {
			response.Error(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	response.JSON(w, http.StatusOK, result)
}

// GetMyKey godoc
// @Summary      Get current user's API key info
// @Tags         auth
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  model.APIKey
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /v1/me/apikey [get]
func (h *AuthHandler) GetMyKey(w http.ResponseWriter, r *http.Request) {
	apiKey := model.APIKeyFromCtx(r.Context())
	if apiKey == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	key, err := h.svc.GetMyKey(r.Context(), apiKey.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	response.JSON(w, http.StatusOK, key)
}
