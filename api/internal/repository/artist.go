package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/florian-alb/music-certifications/internal/model"
)

type ArtistRepo struct {
	db *sql.DB
}

func NewArtistRepo(db *sql.DB) *ArtistRepo {
	return &ArtistRepo{db: db}
}

func (r *ArtistRepo) GetByID(ctx context.Context, id string) (*model.Artist, error) {
	var a model.Artist
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, country, genres, created_at FROM artists WHERE id = $1`, id,
	).Scan(&a.ID, &a.Name, &a.Country, &a.Genres, &a.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get artist: %w", err)
	}
	return &a, nil
}
