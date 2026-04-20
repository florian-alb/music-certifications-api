# Music Certifications API

API REST exposant les certifications musicales (or, platine, diamant…) issues du SNEP, monétisée par clé API avec rate limiting.

## Stack

- **API** — Go 1.22+ (`net/http`, zéro framework)
- **Base de données** — PostgreSQL 16
- **Cache / Rate limiting** — Redis 7
- **Scraper** — Python + Playwright *(à venir)*

## Documentation

L'interface Swagger UI est disponible à l'adresse suivante une fois le serveur lancé :

```
http://localhost:8000/swagger/index.html
```

Pour régénérer la doc après modification des routes ou annotations :

```bash
cd api && ~/go/bin/swag init -g cmd/server/main.go
```

## Démarrage rapide

### 1. Lancer les services

```bash
docker compose up -d
```

### 2. Appliquer les migrations

```bash
./manage.sh migration:run
```

### 3. Lancer l'API

```bash
./manage.sh server
# → Server running on :8000
```

## manage.sh

Script de gestion du projet, à lancer depuis la racine.

```bash
./manage.sh migration:create <name>   # Crée api/migrations/00N_<name>.sql
./manage.sh migration:run             # Applique les migrations en attente
./manage.sh migration:status          # Affiche [x] appliquée / [ ] pending
./manage.sh server                    # Lance go run ./cmd/server
```

Les migrations appliquées sont trackées dans une table `schema_migrations` en base — aucune migration n'est rejouée deux fois.

## Configuration

Variables d'environnement (fichier `api/.env`) :

| Variable | Défaut |
|---|---|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/musiccerts?sslmode=disable` |
| `REDIS_URL` | `redis://localhost:6379` |
| `PORT` | `8000` |

## Endpoints

### Auth (public)

```
POST /v1/auth/register   { "email": "...", "password": "..." }
POST /v1/auth/login      { "email": "...", "password": "..." }
```

L'inscription et la connexion retournent un `plain_key` — c'est votre clé API. **Conservez-la**, elle n'est pas stockée en clair.

### Données (requiert `X-API-Key`)

```
GET /v1/certifications?country=fr&level=platine&page=1&per_page=20
GET /v1/certifications/{id}
GET /v1/artists/{id}/certifications
GET /v1/releases/{id}/certifications
GET /v1/search?q=Jul&country=fr
GET /v1/sources
GET /v1/me/apikey
```

Tous les endpoints protégés exigent le header :

```
X-API-Key: <votre_clé>
```

La réponse inclut `X-RateLimit-Remaining` pour suivre votre quota.

## Tiers & Rate Limiting

| Tier | Limite | Prix |
|---|---|---|
| **Free** | 100 req/jour | 0 € |
| **Starter** | 10 000 req/jour | ~9 €/mois |
| **Pro** | 100 000 req/jour | ~29 €/mois |
| **Enterprise** | Illimité + SLA | Sur devis |

Le compteur est un sliding window Redis par clé API, remis à zéro chaque jour à minuit.

## Exemple

```bash
# Inscription
curl -s -X POST http://localhost:8000/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","password":"secret"}' | jq .

# Recherche
curl -s http://localhost:8000/v1/search?q=Jul&country=fr \
  -H "X-API-Key: <plain_key>" | jq .

# Certifications platine françaises
curl -s "http://localhost:8000/v1/certifications?country=fr&level=platine" \
  -H "X-API-Key: <plain_key>" | jq .
```

## Architecture

```
Request
  → middleware/auth       (valide X-API-Key, injecte dans contexte)
  → middleware/ratelimit  (sliding window Redis)
  → handler              (lecture req, écriture res)
  → service              (logique métier)
  → repository           (SQL uniquement)
  → PostgreSQL
```

```
api/
├── cmd/server/main.go          # Entry point, routing
├── internal/
│   ├── handler/                # HTTP handlers
│   ├── service/                # Logique métier
│   ├── repository/             # Accès PostgreSQL
│   ├── middleware/             # Auth + rate limiting
│   ├── cache/                  # Client Redis
│   └── model/                  # Types partagés
├── pkg/
│   ├── database/               # Connexion PostgreSQL
│   └── response/               # Helpers JSON
└── migrations/
    └── 001_init.sql
```

## Modèle de données

```
Artist      → id, name, country, genres[]
Release     → id, artist_id, title, type, release_date
Certification → id, release_id, source, level, certified_at, sales_equivalent, country
User        → id, email, password_hash
APIKey      → id, user_id, key_hash, tier, req_count, expires_at
ScrapingLog → id, source, run_at, status, records_upserted, error
```
