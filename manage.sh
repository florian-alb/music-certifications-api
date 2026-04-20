#!/usr/bin/env bash
set -euo pipefail

MIGRATIONS_DIR="api/migrations"
DB_USER="postgres"
DB_NAME="musiccerts"

# Exécute psql dans le container postgres
_psql() {
  docker compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" "$@"
}

_ensure_tracking_table() {
  _psql -q -c "
    CREATE TABLE IF NOT EXISTS schema_migrations (
      version    TEXT PRIMARY KEY,
      applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );"
}

_next_number() {
  local last
  last=$(ls "$MIGRATIONS_DIR"/*.sql 2>/dev/null \
    | xargs -I{} basename {} .sql \
    | grep -oE '^[0-9]+' \
    | sort -n \
    | tail -1 || true)
  printf "%03d" $(( 10#${last:-0} + 1 ))
}

case "${1:-help}" in

  migration:create)
    name="${2:-}"
    if [ -z "$name" ]; then
      echo "Usage: ./manage.sh migration:create <name>"
      exit 1
    fi
    file="$MIGRATIONS_DIR/$(_next_number)_${name}.sql"
    touch "$file"
    echo "Created → $file"
    ;;

  migration:run)
    _ensure_tracking_table
    applied=0
    for path in $(ls "$MIGRATIONS_DIR"/*.sql 2>/dev/null | sort); do
      version=$(basename "$path" .sql)
      exists=$(_psql -tAq -c \
        "SELECT 1 FROM schema_migrations WHERE version = '$version';" 2>/dev/null || true)
      if [ "$exists" = "1" ]; then
        echo "  ✓ $version"
      else
        echo "  → $version"
        _psql -q < "$path"
        _psql -q -c "INSERT INTO schema_migrations (version) VALUES ('$version');"
        applied=$((applied + 1))
      fi
    done
    echo "Done — $applied migration(s) applied."
    ;;

  migration:status)
    _ensure_tracking_table
    for path in $(ls "$MIGRATIONS_DIR"/*.sql 2>/dev/null | sort); do
      version=$(basename "$path" .sql)
      exists=$(_psql -tAq -c \
        "SELECT 1 FROM schema_migrations WHERE version = '$version';" 2>/dev/null || true)
      if [ "$exists" = "1" ]; then
        echo "  [x] $version"
      else
        echo "  [ ] $version"
      fi
    done
    ;;

  server)
    cd api && go run ./cmd/server
    ;;

  help|*)
    echo "Usage: ./manage.sh <command>"
    echo ""
    echo "  migration:create <name>   Crée une nouvelle migration"
    echo "  migration:run             Applique les migrations en attente"
    echo "  migration:status          Affiche l'état des migrations"
    echo "  server                    Lance le serveur API"
    ;;

esac
