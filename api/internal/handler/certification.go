package handler

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/florian-alb/music-certifications/internal/model"
	"github.com/florian-alb/music-certifications/pkg/response"
)

type certService interface {
	List(ctx context.Context, country, level string, page, perPage int) ([]model.CertificationFull, int, error)
	GetByID(ctx context.Context, id string) (*model.CertificationFull, error)
	ListByArtist(ctx context.Context, artistID string, page, perPage int) ([]model.CertificationFull, int, error)
	ListByRelease(ctx context.Context, releaseID string) ([]model.CertificationFull, error)
	Search(ctx context.Context, q, country string, page, perPage int) ([]model.CertificationFull, int, error)
	ListSources(ctx context.Context) ([]string, error)
}

type CertificationHandler struct {
	svc certService
}

func NewCertificationHandler(svc certService) *CertificationHandler {
	return &CertificationHandler{svc: svc}
}

// List godoc
// @Summary      List certifications
// @Tags         certifications
// @Produce      json
// @Security     ApiKeyAuth
// @Param        country  query     string  false  "Filter by country code (e.g. FR)"
// @Param        level    query     string  false  "Filter by level (e.g. platinum)"
// @Param        page     query     int     false  "Page number"     default(1)
// @Param        per_page query     int     false  "Items per page"  default(20)
// @Success      200      {object}  object{data=[]model.CertificationFull,meta=object}
// @Failure      500      {object}  response.ErrorResponse
// @Router       /v1/certifications [get]
func (h *CertificationHandler) List(w http.ResponseWriter, r *http.Request) {
	country := r.URL.Query().Get("country")
	level := r.URL.Query().Get("level")
	page, perPage := pageParams(r)

	certs, total, err := h.svc.List(r.Context(), country, level, page, perPage)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	response.JSON(w, http.StatusOK, paginated(certs, page, perPage, total))
}

// GetByID godoc
// @Summary      Get a certification by ID
// @Tags         certifications
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "Certification ID"
// @Success      200  {object}  model.CertificationFull
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /v1/certifications/{id} [get]
func (h *CertificationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cert, err := h.svc.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.Error(w, http.StatusNotFound, "certification not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	response.JSON(w, http.StatusOK, cert)
}

// ListByArtist godoc
// @Summary      List certifications for an artist
// @Tags         certifications
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id       path      string  true   "Artist ID"
// @Param        page     query     int     false  "Page number"    default(1)
// @Param        per_page query     int     false  "Items per page" default(20)
// @Success      200      {object}  object{data=[]model.CertificationFull,meta=object}
// @Failure      500      {object}  response.ErrorResponse
// @Router       /v1/artists/{id}/certifications [get]
func (h *CertificationHandler) ListByArtist(w http.ResponseWriter, r *http.Request) {
	page, perPage := pageParams(r)
	certs, total, err := h.svc.ListByArtist(r.Context(), r.PathValue("id"), page, perPage)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	response.JSON(w, http.StatusOK, paginated(certs, page, perPage, total))
}

// ListByRelease godoc
// @Summary      List certifications for a release
// @Tags         certifications
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "Release ID"
// @Success      200  {object}  object{data=[]model.CertificationFull}
// @Failure      500  {object}  response.ErrorResponse
// @Router       /v1/releases/{id}/certifications [get]
func (h *CertificationHandler) ListByRelease(w http.ResponseWriter, r *http.Request) {
	certs, err := h.svc.ListByRelease(r.Context(), r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"data": certs})
}

// Search godoc
// @Summary      Search certifications
// @Tags         certifications
// @Produce      json
// @Security     ApiKeyAuth
// @Param        q        query     string  true   "Search query (artist or release name)"
// @Param        country  query     string  false  "Filter by country code"
// @Param        page     query     int     false  "Page number"    default(1)
// @Param        per_page query     int     false  "Items per page" default(20)
// @Success      200      {object}  object{data=[]model.CertificationFull,meta=object}
// @Failure      400      {object}  response.ErrorResponse
// @Failure      500      {object}  response.ErrorResponse
// @Router       /v1/search [get]
func (h *CertificationHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		response.Error(w, http.StatusBadRequest, "q is required")
		return
	}
	page, perPage := pageParams(r)
	certs, total, err := h.svc.Search(r.Context(), q, r.URL.Query().Get("country"), page, perPage)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	response.JSON(w, http.StatusOK, paginated(certs, page, perPage, total))
}

// ListSources godoc
// @Summary      List available certification sources
// @Tags         certifications
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  object{data=[]string}
// @Failure      500  {object}  response.ErrorResponse
// @Router       /v1/sources [get]
func (h *CertificationHandler) ListSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.svc.ListSources(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"data": sources})
}

func pageParams(r *http.Request) (page, perPage int) {
	page = queryInt(r, "page", 1)
	perPage = queryInt(r, "per_page", 20)
	return
}

func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return def
	}
	return n
}

func paginated(data any, page, perPage, total int) map[string]any {
	return map[string]any{
		"data": data,
		"meta": map[string]any{
			"page":     page,
			"per_page": perPage,
			"total":    total,
		},
	}
}
