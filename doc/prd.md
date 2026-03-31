# Cahier des charges — API Certifications Musicales v1.2

## Compréhension du besoin

Construire une **API monétisable** qui expose des données sur les certifications musicales (or, platine, diamant…), en commençant par la France (SNEP), avec une ambition d'extension internationale. Les données sont collectées par scraping périodique et stockées en base pour être servies à forte charge.

---

## Stack technique

| Brique          | Choix                             | Raison                                                           |
| --------------- | --------------------------------- | ---------------------------------------------------------------- |
| **API**         | Go 1.22+ natif (`net/http`)       | Zéro dépendance, routing paramétré natif, performances maximales |
| **BDD**         | PostgreSQL                        | Robuste, full-text search, JSONB, éprouvé                        |
| **Cache**       | Redis                             | Réponses cachées, rate limiting par API key                      |
| **Scraping**    | Python + Playwright + APScheduler | JS-heavy, isolé du core API                                      |
| **Déploiement** | Docker Compose                    | Simple, portable, scalable vers K8s                              |

Les deux services sont **découplés** — le scraper Python et l'API Go partagent uniquement PostgreSQL et Redis.

```
Scraper (Python)              API (Go natif)
      ↓                             ↓
  PostgreSQL   ←—————————→    PostgreSQL
      ↓                             ↓
   Redis        ←—————————→     Redis
```

---

## Modèle de données

```
Artist         → id, name, country, genres[]
Release        → id, artist_id, title, type (single/album), release_date
Certification  → id, release_id, source (SNEP/RIAA…), level (or/platine/diamant…),
                 certified_at, sales_equivalent, country, raw_data (JSONB)
APIKey         → id, user_id, key_hash, tier, req_count, created_at, expires_at
ScrapingLog    → id, source, run_at, status, records_upserted, error
```

---

## Endpoints cibles (v1)

```
# Certifications
GET  /v1/certifications?country=fr&level=platine&page=1
GET  /v1/certifications/{id}
GET  /v1/artists/{id}/certifications
GET  /v1/releases/{id}/certifications
GET  /v1/search?q=Jul&country=fr
GET  /v1/sources

# Auth
POST /v1/auth/register
POST /v1/auth/login
GET  /v1/me/apikey
```

---

## Monétisation

| Tier           | Limite         | Prix      |
| -------------- | -------------- | --------- |
| **Free**       | 100 req/jour   | 0€        |
| **Starter**    | 10k req/jour   | ~9€/mois  |
| **Pro**        | 100k req/jour  | ~29€/mois |
| **Enterprise** | Illimité + SLA | Sur devis |

Rate limiting via **Redis sliding window** par API key, implémenté manuellement en Go — une vingtaine de lignes.

---

## Flow scraping SNEP

```
Scheduler (2x/semaine)
  → Playwright scrape snepmusique.com
  → Parse & normalise les données
  → Upsert PostgreSQL (pas de doublons)
  → Invalidation cache Redis
  → Log dans ScrapingLog
```

---

## Ce qu'on ne fait PAS en v1

- Internationalisation (RIAA, BPI…) → v2
- Dashboard admin
- Billing Stripe → v2
- Queue Celery (scraping simple suffit pour 1 source)

---
