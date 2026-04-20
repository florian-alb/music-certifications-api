package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/florian-alb/music-certifications/internal/model"
)

type CertificationRepo struct {
	db *sql.DB
}

func NewCertificationRepo(db *sql.DB) *CertificationRepo {
	return &CertificationRepo{db: db}
}

const certBase = `
	SELECT
		c.id,
		a.id, a.name,
		r.id, r.title, r.type,
		c.source, c.level, c.certified_at, c.sales_equivalent, c.country, c.created_at,
		COUNT(*) OVER() AS total
	FROM certifications c
	JOIN releases r ON r.id = c.release_id
	JOIN artists  a ON a.id = r.artist_id`

func scanRow(rows *sql.Rows) (model.CertificationFull, int, error) {
	var c model.CertificationFull
	var total int
	var certAt sql.NullTime
	var sales sql.NullInt32
	if err := rows.Scan(
		&c.ID,
		&c.Artist.ID, &c.Artist.Name,
		&c.Release.ID, &c.Release.Title, &c.Release.Type,
		&c.Source, &c.Level, &certAt, &sales, &c.Country, &c.CreatedAt,
		&total,
	); err != nil {
		return c, 0, err
	}
	if certAt.Valid {
		c.CertifiedAt = &certAt.Time
	}
	if sales.Valid {
		v := int(sales.Int32)
		c.SalesEquivalent = &v
	}
	return c, total, nil
}

func collect(rows *sql.Rows) ([]model.CertificationFull, int, error) {
	defer rows.Close()
	var out []model.CertificationFull
	var total int
	for rows.Next() {
		cert, t, err := scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, cert)
		total = t
	}
	if out == nil {
		out = []model.CertificationFull{}
	}
	return out, total, rows.Err()
}

func (r *CertificationRepo) List(ctx context.Context, country, level string, page, perPage int) ([]model.CertificationFull, int, error) {
	q := certBase + `
	WHERE ($1::text = '' OR c.country = $1)
	  AND ($2::text = '' OR c.level   = $2)
	ORDER BY c.certified_at DESC NULLS LAST
	LIMIT $3 OFFSET $4`

	rows, err := r.db.QueryContext(ctx, q, country, level, perPage, (page-1)*perPage)
	if err != nil {
		return nil, 0, fmt.Errorf("list certifications: %w", err)
	}
	return collect(rows)
}

func (r *CertificationRepo) GetByID(ctx context.Context, id string) (*model.CertificationFull, error) {
	rows, err := r.db.QueryContext(ctx, certBase+` WHERE c.id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("get certification: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	cert, _, err := scanRow(rows)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

func (r *CertificationRepo) ListByArtist(ctx context.Context, artistID string, page, perPage int) ([]model.CertificationFull, int, error) {
	q := certBase + `
	WHERE a.id = $1
	ORDER BY c.certified_at DESC NULLS LAST
	LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, q, artistID, perPage, (page-1)*perPage)
	if err != nil {
		return nil, 0, fmt.Errorf("list by artist: %w", err)
	}
	return collect(rows)
}

func (r *CertificationRepo) ListByRelease(ctx context.Context, releaseID string) ([]model.CertificationFull, error) {
	q := certBase + `
	WHERE r.id = $1
	ORDER BY c.certified_at DESC NULLS LAST`

	rows, err := r.db.QueryContext(ctx, q, releaseID)
	if err != nil {
		return nil, fmt.Errorf("list by release: %w", err)
	}
	certs, _, err := collect(rows)
	return certs, err
}

func (r *CertificationRepo) Search(ctx context.Context, q, country string, page, perPage int) ([]model.CertificationFull, int, error) {
	query := certBase + `
	WHERE (a.name ILIKE '%' || $1 || '%' OR r.title ILIKE '%' || $1 || '%')
	  AND ($2::text = '' OR c.country = $2)
	ORDER BY c.certified_at DESC NULLS LAST
	LIMIT $3 OFFSET $4`

	rows, err := r.db.QueryContext(ctx, query, q, country, perPage, (page-1)*perPage)
	if err != nil {
		return nil, 0, fmt.Errorf("search: %w", err)
	}
	return collect(rows)
}

func (r *CertificationRepo) ListSources(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT DISTINCT source FROM certifications ORDER BY source`)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}
	defer rows.Close()
	var sources []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}
	if sources == nil {
		sources = []string{}
	}
	return sources, rows.Err()
}
