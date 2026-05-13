# Database

## Overview

PostgreSQL 16 is the single source of truth. All writes go through catalog-api.

## Tables

### publishers
Represents a marketplace seller. Identified by `(source, source_publisher_id)`.

| Column | Type | Description |
|--------|------|-------------|
| id | uuid | Primary key |
| source | text | e.g. `facebook_marketplace` |
| source_publisher_id | text | Publisher ID from the source |
| display_name | text | Display name as seen on listing card |
| profile_url | text | Full profile URL |
| profile_url_hash | text | SHA-256 of profile_url for dedup |
| location_hint | text | Location text from listing card |
| raw_json | jsonb | Raw data from the source |
| first_seen_at | timestamptz | First time this publisher was observed |
| last_seen_at | timestamptz | Most recent observation |

### listings
The canonical record for a marketplace item. Identified by `(source, source_listing_id)`.

| Column | Type | Description |
|--------|------|-------------|
| id | uuid | Primary key |
| source | text | e.g. `facebook_marketplace` |
| source_listing_id | text | Listing ID from the source |
| publisher_id | uuid | FK → publishers |
| canonical_url | text | Canonical listing URL |
| canonical_url_hash | text | SHA-256 of canonical_url |
| title_current | text | Most recent title |
| price_current | numeric(14,2) | Most recent price |
| currency | text | Currency code (e.g. `THB`) |
| location_current | text | Most recent location text |
| status | text | `active`, `sold`, `removed` |
| detail_status | text | `not_requested`, `requested`, `captured` |
| enrichment_status | text | `not_requested`, `requested`, `enriched`, `failed` |
| duplicate_status | text | `unique`, `candidate`, `duplicate` |
| product_cluster_id | uuid | FK → product_clusters (nullable) |
| first_seen_at | timestamptz | First observation time |
| last_seen_at | timestamptz | Most recent observation time |
| version | bigint | Incremented on each update |
| raw_json | jsonb | Raw data from the source |

### search_queries
User-defined search configurations that drive the scheduler.

| Column | Type | Description |
|--------|------|-------------|
| id | uuid | Primary key |
| name | text | Human-readable name |
| query | text | Search string |
| source | text | Marketplace source |
| location_hint | text | Optional location filter |
| enabled | boolean | Whether the query is active |
| max_cards_per_run | int | Max listing cards per scrape |
| max_scrolls_per_run | int | Max scroll actions per scrape |
| active_window_start | time | Only run after this time (local) |
| active_window_end | time | Only run before this time (local) |
| timezone | text | Timezone for active window |
| interval_minutes | int | Target interval between runs |
| jitter_minutes | int | Random ± jitter added to interval |
| skip_probability | numeric(5,4) | Probability of skipping a scheduled run |
| max_runs_per_day | int | Hard cap on daily runs |

### search_sessions
Tracks each individual scrape run triggered by a `SearchRequested` event.

| Column | Type | Description |
|--------|------|-------------|
| id | uuid | Primary key (= event's search_session_id) |
| search_query_id | uuid | FK → search_queries |
| status | text | `created`, `running`, `completed`, `failed` |
| started_at | timestamptz | When grabber began |
| finished_at | timestamptz | When grabber finished |
| error_code | text | Error code on failure |
| error_message | text | Error message on failure |

### listing_observations
Append-only log of every time a listing was seen. Never updated.

| Column | Type | Description |
|--------|------|-------------|
| id | uuid | Primary key |
| listing_id | uuid | FK → listings |
| search_session_id | uuid | FK → search_sessions |
| search_query_id | uuid | FK → search_queries |
| observed_at | timestamptz | When the card was seen |
| position | int | Position in the search results |
| title_observed | text | Title at observation time |
| price_observed | numeric(14,2) | Price at observation time |
| price_text | text | Raw price string |
| currency | text | Currency code |
| location_observed | text | Location at observation time |
| url_observed | text | URL at observation time |
| raw_json | jsonb | Full raw card data |

### listing_detail_snapshots
Stores full detail page captures (future use).

### listing_images
Stores image metadata per listing, used for deduplication and pHash.

### product_clusters
Groups similar listings into a product cluster (future: similarity-service).

### duplicate_candidates
Records potential duplicates between two listings (future: similarity-service).

### listing_metadata
Enriched structured metadata per listing (future: enricher).

| Column | Type | Description |
|--------|------|-------------|
| listing_id | uuid | PK + FK → listings |
| language | text | Detected language code |
| normalized_title_en | text | Normalized English title |
| normalized_description_en | text | Normalized English description |
| category | text | Top-level category |
| subcategory | text | Subcategory |
| manufacturer | text | Brand / manufacturer |
| model | text | Product model |
| specs_json | jsonb | Extracted specs (RAM, storage, etc.) |
| confidence_json | jsonb | Confidence scores per field |
| llm_model | text | LLM model used for extraction |

### price_observations
Append-only log of every price seen for a listing.

| Column | Type | Description |
|--------|------|-------------|
| id | uuid | Primary key |
| listing_id | uuid | FK → listings |
| observed_at | timestamptz | When the price was seen |
| price_amount | numeric(14,2) | Price value |
| currency | text | Currency code |
| source | text | Source of observation |

### market_price_baselines
Aggregated market price statistics per product type (future: scoring-service).

### listing_scores
Computed scores per listing (future: scoring-service).

| Column | Type | Description |
|--------|------|-------------|
| listing_id | uuid | PK + FK → listings |
| deal_score | numeric(5,4) | Composite deal quality score (0–1) |
| price_score | numeric(5,4) | Price vs market baseline (0–1) |
| quality_score | numeric(5,4) | Listing completeness score (0–1) |
| risk_score | numeric(5,4) | Risk indicator (0–1, higher = riskier) |
| freshness_score | numeric(5,4) | How recently observed (0–1) |
| fit_score | numeric(5,4) | Match to user preferences (0–1) |
| fair_price_estimate | numeric(14,2) | Estimated fair market price |
| discount_ratio | numeric(8,4) | (fair - current) / fair |
| decision | text | `buy`, `watch`, `skip` |

### llm_requests
Audit log of all LLM API calls, with input hash for caching (future: llm-gateway).

### task_audit_log
Audit log of all service actions (event consumed, status, error).

## Key Design Decisions

### Upsert pattern
`listings` and `publishers` use `INSERT ... ON CONFLICT DO UPDATE` with
`(source, source_listing_id)` as the conflict target. This ensures idempotency
when the same listing is observed multiple times.

### Append-only observations
`listing_observations` and `price_observations` are never updated. Each
observation is a new row. This preserves full history.

### Version column
`listings.version` is incremented on each update. Useful for detecting
concurrent modifications and for downstream consumers to check if state changed.

### updated_at trigger
All tables with `updated_at` use a PostgreSQL trigger (`set_updated_at`) to
automatically update the timestamp on every row change.
