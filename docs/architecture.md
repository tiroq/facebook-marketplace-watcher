# Architecture

## Overview

fb-market-watcher is a personal Facebook Marketplace intelligence system
built as a monorepo with Go backend services and a TypeScript browser grabber.

## Service Responsibilities

### catalog-api (Go)
- **Owns all database writes**
- Exposes HTTP API for listings, search queries, and observations
- Consumes `fb.listing.card.observed` from NATS
- Atomically upserts: publisher → listing → listing_observation → price_observation
- Publishes `fb.listing.upserted` after successful upsert

### scheduler (Go)
- Emits `fb.search.requested` events on a configurable schedule
- Applies skip probability and jitter to behave as a low-frequency personal tool
- Supports active window (only run during certain hours)
- Supports dry-run mode for testing

### grabber (TypeScript)
- **Browser executor only — not a decision service**
- Subscribes to `fb.search.requested`
- Opens Facebook Marketplace in a persistent browser context
- Scrolls and extracts visible listing cards
- Publishes `fb.listing.card.observed` for each card found
- Stops and reports if login/checkpoint/captcha is detected

## Data Flow

```
scheduler
  → fb.search.requested
    → grabber
      → fb.search.started
      → [browser: search + scroll + extract]
      → fb.listing.card.observed (×N)
      → fb.search.completed
        → catalog-api
          → upsert publisher
          → upsert listing
          → insert listing_observation
          → insert price_observation
          → fb.listing.upserted
```

## Catalog API Owns Writes

All data persistence goes through catalog-api. Other services never write
to the database directly. This ensures:
- Single source of truth for data integrity
- Consistent upsert logic
- Atomic transactions
- Audit trail via listing_observations

## Grabber as Executor

The grabber makes no decisions about what to search for or how to process results.
It receives a task, executes it in the browser, and publishes raw observations.
All business logic lives in catalog-api or scheduler.

## Future Service Boundaries

| Service | Boundary |
|---------|----------|
| planner | Decides what to search for based on user goals |
| llm-gateway | Proxies LLM calls with caching and rate limiting |
| enricher | Normalizes and enriches listing metadata |
| similarity-service | Detects duplicates and clusters products |
| scoring-service | Scores listings for deal quality |
| notifier | Sends alerts for interesting listings |

## NATS JetStream

All services communicate via NATS JetStream with:
- Stream: `FB_EVENTS`
- Subjects: `fb.>`
- Storage: File (persistent across restarts)
- Retention: Limits (configurable)

## PostgreSQL

PostgreSQL 16 is the single source of truth. Tables follow these principles:
- `source` + `source_listing_id` uniquely identifies a listing
- `publishers` are separate from listings
- Observations are append-only
- Listings store current state
- Price observations are append-only
