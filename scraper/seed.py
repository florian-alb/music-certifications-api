#!/usr/bin/env python3
"""Seed the database from baserow_rows.json."""

import json
import os
import sys
from datetime import datetime

import psycopg2
from psycopg2.extras import RealDictCursor

DB_URL = os.getenv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/musiccerts")

VALID_CATEGORIES = {"Singles", "Single", "Albums", "Vidéos"}
VALID_CERTIFICATIONS = {
    "Or", "Platine", "Double Platine", "Triple Platine",
    "Diamant", "Double Diamant", "Double diamant", "Double or",
    "Triple Diamant", "Triple diamant", "Quadruple Diamant",
}

CATEGORY_MAP = {
    "Singles": "single",
    "Single": "single",
    "Albums": "album",
    "Vidéos": "video",
}

LEVEL_NORMALIZE = {k.lower(): k for k in VALID_CERTIFICATIONS}


def parse_date(s: str) -> datetime | None:
    if not s or s.strip() == "":
        return None
    try:
        return datetime.strptime(s.strip(), "%d/%m/%Y")
    except ValueError:
        return None


def upsert_artist(cur, name: str) -> str:
    cur.execute(
        """
        INSERT INTO artists (name, country)
        VALUES (%s, 'fr')
        ON CONFLICT (name, country) DO NOTHING
        RETURNING id
        """,
        (name,),
    )
    row = cur.fetchone()
    if row:
        return row["id"]
    cur.execute("SELECT id FROM artists WHERE name = %s AND country = 'fr'", (name,))
    return cur.fetchone()["id"]


def upsert_release(cur, artist_id: str, title: str, release_type: str, release_date) -> str:
    cur.execute(
        """
        INSERT INTO releases (artist_id, title, type, release_date)
        VALUES (%s, %s, %s, %s)
        ON CONFLICT (artist_id, title, type) DO NOTHING
        RETURNING id
        """,
        (artist_id, title, release_type, release_date),
    )
    row = cur.fetchone()
    if row:
        return row["id"]
    cur.execute(
        "SELECT id FROM releases WHERE artist_id = %s AND title = %s AND type = %s",
        (artist_id, title, release_type),
    )
    return cur.fetchone()["id"]


def upsert_certification(cur, release_id: str, level: str, certified_at, raw: dict):
    cur.execute(
        """
        INSERT INTO certifications (release_id, source, level, certified_at, country, raw_data)
        VALUES (%s, 'snep', %s, %s, 'fr', %s)
        ON CONFLICT (release_id, source) DO UPDATE
            SET level        = EXCLUDED.level,
                certified_at = EXCLUDED.certified_at,
                raw_data     = EXCLUDED.raw_data,
                updated_at   = NOW()
        """,
        (release_id, level, certified_at, json.dumps(raw, ensure_ascii=False)),
    )


def main():
    json_path = os.path.join(os.path.dirname(__file__), "baserow_rows.json")
    if not os.path.exists(json_path):
        json_path = os.path.join(os.path.dirname(__file__), "..", "baserow_rows.json")
    with open(json_path, encoding="utf-8") as f:
        rows = json.load(f)

    conn = psycopg2.connect(DB_URL)
    conn.autocommit = False

    skipped = 0
    inserted = 0

    with conn.cursor(cursor_factory=RealDictCursor) as cur:
        for row in rows:
            category_raw = row.get("Catégorie", "")
            cert_raw = row.get("Certification", "")

            if category_raw not in VALID_CATEGORIES:
                print(f"[SKIP] id={row['id']} — catégorie invalide: {category_raw!r}")
                skipped += 1
                continue

            normalized_cert = LEVEL_NORMALIZE.get(cert_raw.lower())
            if not normalized_cert:
                print(f"[SKIP] id={row['id']} — certification invalide: {cert_raw!r}")
                skipped += 1
                continue

            artist_name = row["Interprete"].strip()
            title = row["Titre"].strip()
            release_type = CATEGORY_MAP[category_raw]
            release_date = parse_date(row.get("Date de sortie", ""))
            certified_at = parse_date(row.get("Date de constat", ""))

            artist_id = upsert_artist(cur, artist_name)
            release_id = upsert_release(cur, artist_id, title, release_type, release_date)
            upsert_certification(cur, release_id, normalized_cert, certified_at, row)
            inserted += 1

        conn.commit()

    conn.close()
    print(f"\nDone — {inserted} upserted, {skipped} skipped.")


if __name__ == "__main__":
    main()
