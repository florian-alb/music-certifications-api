package model

import "time"

type Artist struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Country   string    `json:"country"`
	Genres    []string  `json:"genres"`
	CreatedAt time.Time `json:"created_at"`
}

type ArtistRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Release struct {
	ID          string     `json:"id"`
	ArtistID    string     `json:"artist_id"`
	Title       string     `json:"title"`
	Type        string     `json:"type"`
	ReleaseDate *time.Time `json:"release_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ReleaseRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
}
