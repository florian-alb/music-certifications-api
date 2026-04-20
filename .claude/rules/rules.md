# Rules — API Certifications Musicales

## Architecture de fichiers

```
music-certifications/
│
├── api/                          # Go
│   ├── cmd/
│   │   └── server/
│   │       └── main.go           # Entry point
│   │
│   ├── internal/
│   │   ├── handler/              # HTTP handlers (lecture req, écriture res)
│   │   │   ├── certification.go
│   │   │   ├── artist.go
│   │   │   ├── search.go
│   │   │   └── auth.go
│   │   │
│   │   ├── service/              # Logique métier
│   │   │   ├── certification.go
│   │   │   ├── artist.go
│   │   │   ├── search.go
│   │   │   └── auth.go
│   │   │
│   │   ├── repository/           # Accès BDD uniquement
│   │   │   ├── certification.go
│   │   │   ├── artist.go
│   │   │   └── auth.go
│   │   │
│   │   ├── middleware/
│   │   │   ├── auth.go           # Validation API key
│   │   │   ├── ratelimit.go      # Redis sliding window
│   │   │   └── logger.go
│   │   │
│   │   ├── cache/
│   │   │   └── redis.go
│   │   │
│   │   └── model/
│   │       ├── certification.go
│   │       ├── artist.go
│   │       └── apikey.go
│   │
│   ├── pkg/
│   │   ├── database/
│   │   │   └── postgres.go
│   │   └── response/
│   │       └── json.go
│   │
│   ├── migrations/
│   │   ├── 001_create_artists.sql
│   │   ├── 002_create_releases.sql
│   │   ├── 003_create_certifications.sql
│   │   └── 004_create_apikeys.sql
│   │
│   ├── go.mod
│   └── go.sum
│
├── scraper/                      # Python
│   ├── scrapers/
│   │   └── snep.py
│   ├── parsers/
│   │   └── snep.py
│   ├── db/
│   │   └── upsert.py
│   ├── scheduler.py
│   ├── requirements.txt
│   └── .env
│
├── docker/
│   ├── api.Dockerfile
│   └── scraper.Dockerfile
│
├── docker-compose.yml
├── .env.example
└── README.md
```

---

## Règles

### Couches Go — ordre strict
```
Request → middleware → handler → service → repository → PostgreSQL
```
- Un handler n'écrit jamais de SQL
- Un repository ne connaît pas Redis
- Chaque couche ne connaît que celle d'en dessous

### internal/
Tout ce qui est dans `internal/` ne peut pas être importé depuis l'extérieur du module.

### pkg/
Contient uniquement ce qui est générique et réutilisable (connexion PG, helpers JSON).

### Scraper
Le scraper est complètement isolé — il ne connaît pas l'API, il parle uniquement à PostgreSQL et Redis.