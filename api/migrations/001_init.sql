CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email        TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE api_keys (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash   TEXT NOT NULL UNIQUE,
    tier       TEXT NOT NULL DEFAULT 'free',
    req_count  BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_user_id  ON api_keys(user_id);

CREATE TABLE artists (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    country    TEXT NOT NULL DEFAULT 'fr',
    genres     TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_artists_name_trgm ON artists USING gin(to_tsvector('simple', name));

CREATE TABLE releases (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    artist_id    UUID NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    title        TEXT NOT NULL,
    type         TEXT NOT NULL DEFAULT 'single',
    release_date DATE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_releases_artist_id ON releases(artist_id);

CREATE TABLE certifications (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    release_id       UUID NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    source           TEXT NOT NULL DEFAULT 'snep',
    level            TEXT NOT NULL,
    certified_at     DATE,
    sales_equivalent INTEGER,
    country          TEXT NOT NULL DEFAULT 'fr',
    raw_data         JSONB,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_certifications_release_id ON certifications(release_id);
CREATE INDEX idx_certifications_country    ON certifications(country);
CREATE INDEX idx_certifications_level      ON certifications(level);
CREATE INDEX idx_certifications_source     ON certifications(source);

CREATE TABLE scraping_logs (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source           TEXT NOT NULL,
    run_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status           TEXT NOT NULL,
    records_upserted INTEGER NOT NULL DEFAULT 0,
    error            TEXT
);
