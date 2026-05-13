# MVP Definition

## What is Implemented

### Data Collection Pipeline
- [x] Scheduler emits `SearchRequested` events with jitter and skip probability
- [x] Grabber subscribes to `SearchRequested` and runs Playwright searches
- [x] Grabber extracts listing cards from Facebook Marketplace
- [x] Grabber publishes `ListingCardObserved` for each card
- [x] Catalog API consumes `ListingCardObserved` and upserts data atomically

### Catalog API
- [x] `GET /healthz` — health check
- [x] `GET /readyz` — readiness check
- [x] `GET /v1/listings` — list listings with filters
- [x] `GET /v1/listings/{id}` — get single listing
- [x] `POST /v1/listings/observations` — manual observation submission
- [x] `GET /v1/search-queries` — list search queries
- [x] `POST /v1/search-queries` — create search query

### Data Storage
- [x] PostgreSQL with all 15 tables
- [x] Publisher upsert by source + source_publisher_id
- [x] Listing upsert by source + source_listing_id
- [x] Listing observation insert (append-only)
- [x] Price observation insert when price is present

### Admin UI
- [x] NocoDB connected to PostgreSQL (browse and edit listings)
- [x] Metabase connected to PostgreSQL (analytics and filtering)

## Intentionally Not Implemented

### Phase 2: Enrichment
- [ ] Language detection
- [ ] Translation to English
- [ ] Category classification
- [ ] Metadata extraction
- [ ] llm-gateway integration

### Phase 3: Intelligence
- [ ] Similarity detection
- [ ] Duplicate candidates
- [ ] Product clusters
- [ ] Scoring (deal score, price score, etc.)
- [ ] Market price baselines

### Phase 4: Notifications
- [ ] Telegram alerts
- [ ] Price drop detection
- [ ] Risky deal flagging

### Infrastructure
- [ ] Planner service (dynamic search query management)
- [ ] Detail page scraping
- [ ] Image downloading and phash

## Milestones

### M1: Foundation (current)
- Monorepo structure
- Docker Compose with all services
- SQL migration
- Basic data collection pipeline
- NocoDB + Metabase UI

### M2: Enrichment
- Text normalization
- Language detection
- Category rules

### M3: Intelligence
- Similarity detection
- Scoring service
- Market baselines

### M4: Notifications
- Telegram bot
- Deal alerts
- Price drop alerts
