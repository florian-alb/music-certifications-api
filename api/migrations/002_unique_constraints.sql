ALTER TABLE artists
    ADD CONSTRAINT uq_artists_name_country UNIQUE (name, country);

ALTER TABLE releases
    ADD CONSTRAINT uq_releases_artist_title_type UNIQUE (artist_id, title, type);

ALTER TABLE certifications
    ADD CONSTRAINT uq_certifications_release_source UNIQUE (release_id, source);
