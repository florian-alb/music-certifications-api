package service

import (
	"context"

	"github.com/florian-alb/music-certifications/internal/model"
)

type certRepo interface {
	List(ctx context.Context, country, level string, page, perPage int) ([]model.CertificationFull, int, error)
	GetByID(ctx context.Context, id string) (*model.CertificationFull, error)
	ListByArtist(ctx context.Context, artistID string, page, perPage int) ([]model.CertificationFull, int, error)
	ListByRelease(ctx context.Context, releaseID string) ([]model.CertificationFull, error)
	Search(ctx context.Context, q, country string, page, perPage int) ([]model.CertificationFull, int, error)
	ListSources(ctx context.Context) ([]string, error)
}

type CertificationService struct {
	repo certRepo
}

func NewCertificationService(repo certRepo) *CertificationService {
	return &CertificationService{repo: repo}
}

func clampPage(page, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return page, perPage
}

func (s *CertificationService) List(ctx context.Context, country, level string, page, perPage int) ([]model.CertificationFull, int, error) {
	page, perPage = clampPage(page, perPage)
	return s.repo.List(ctx, country, level, page, perPage)
}

func (s *CertificationService) GetByID(ctx context.Context, id string) (*model.CertificationFull, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CertificationService) ListByArtist(ctx context.Context, artistID string, page, perPage int) ([]model.CertificationFull, int, error) {
	page, perPage = clampPage(page, perPage)
	return s.repo.ListByArtist(ctx, artistID, page, perPage)
}

func (s *CertificationService) ListByRelease(ctx context.Context, releaseID string) ([]model.CertificationFull, error) {
	return s.repo.ListByRelease(ctx, releaseID)
}

func (s *CertificationService) Search(ctx context.Context, q, country string, page, perPage int) ([]model.CertificationFull, int, error) {
	page, perPage = clampPage(page, perPage)
	return s.repo.Search(ctx, q, country, page, perPage)
}

func (s *CertificationService) ListSources(ctx context.Context) ([]string, error) {
	return s.repo.ListSources(ctx)
}
