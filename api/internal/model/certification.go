package model

import "time"

type CertificationFull struct {
	ID              string     `json:"id"`
	Artist          ArtistRef  `json:"artist"`
	Release         ReleaseRef `json:"release"`
	Source          string     `json:"source"`
	Level           string     `json:"level"`
	CertifiedAt     *time.Time `json:"certified_at,omitempty"`
	SalesEquivalent *int       `json:"sales_equivalent,omitempty"`
	Country         string     `json:"country"`
	CreatedAt       time.Time  `json:"created_at"`
}

type ScrapingLog struct {
	ID              string    `json:"id"`
	Source          string    `json:"source"`
	RunAt           time.Time `json:"run_at"`
	Status          string    `json:"status"`
	RecordsUpserted int       `json:"records_upserted"`
	Error           *string   `json:"error,omitempty"`
}
