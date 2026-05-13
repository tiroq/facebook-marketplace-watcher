# Future Services

This document describes services planned for phases after the MVP.
All services listed here are **not implemented**. See individual service
README files for details.

## Service Map

```
                    ┌─────────────┐
                    │  scheduler  │
                    └──────┬──────┘
                           │ fb.search.requested
                    ┌──────▼──────┐
                    │   grabber   │
                    └──────┬──────┘
                           │ fb.listing.card.observed
                    ┌──────▼──────┐
                    │ catalog-api │ ◄── all DB writes
                    └──────┬──────┘
                           │ fb.listing.upserted
          ┌────────────────┼────────────────┐
          │                │                │
   ┌──────▼──────┐  ┌──────▼───────┐  ┌────▼──────────────┐
   │   enricher  │  │  similarity  │  │  scoring-service   │
   └──────┬──────┘  │   -service   │  └────┬──────────────┘
          │         └──────┬───────┘       │ fb.listing.score.updated
          │ fb.listing     │               │
          │ .enrichment    │               ┌▼────────┐
          │ .requested     │               │notifier │
   ┌──────▼──────┐         │               └─────────┘
   │ llm-gateway │         │
   └─────────────┘         │
                           │
                    (product_clusters,
                     duplicate_candidates)
```

## Phase 2: Enrichment

### enricher
Transforms raw listing text into structured, normalized metadata.

- **Status**: Not implemented
- **README**: [apps/enricher/README.md](../apps/enricher/README.md)
- **Triggers**: `fb.listing.upserted` where `enrichment_status = not_requested`
- **Produces**: `fb.listing.enrichment.requested`
- **Updates**: `listing_metadata`, `listings.enrichment_status`

### llm-gateway
Central proxy for all LLM API calls with caching and rate limiting.

- **Status**: Not implemented
- **README**: [apps/llm-gateway/README.md](../apps/llm-gateway/README.md)
- **Consumes**: `fb.listing.enrichment.requested`
- **Produces**: `fb.listing.enriched`, `fb.ops.error`
- **Audit**: `llm_requests` table with `input_hash` for caching

## Phase 3: Intelligence

### similarity-service
Detects duplicate listings and groups similar products into clusters.

- **Status**: Not implemented
- **README**: [apps/similarity-service/README.md](../apps/similarity-service/README.md)
- **Consumes**: `fb.listing.similarity.requested`
- **Produces**: `fb.listing.similarity.completed`, `fb.listing.duplicate_candidate.created`
- **Updates**: `duplicate_candidates`, `product_clusters`, `listings.duplicate_status`

### scoring-service
Evaluates each listing and assigns deal quality scores.

- **Status**: Not implemented
- **README**: [apps/scoring-service/README.md](../apps/scoring-service/README.md)
- **Consumes**: `fb.listing.score.requested`
- **Produces**: `fb.listing.score.updated`
- **Updates**: `listing_scores`

## Phase 4: Notifications

### notifier
Sends Telegram alerts for high-score listings and price drops.

- **Status**: Not implemented
- **README**: [apps/notifier/README.md](../apps/notifier/README.md)
- **Consumes**: `fb.listing.score.updated`, `fb.alert.candidate.created`
- **Produces**: Telegram messages via Bot API

## Planned but Not Scoped

### planner
Dynamically manages search queries based on user goals and market conditions.
Not yet designed.

### detail-scraper
Scrapes individual listing detail pages for full descriptions, images, and
seller information. Would publish `fb.listing.detail.observed` and update
`listing_detail_snapshots`.

## Implementation Priority

The recommended implementation order follows data dependency:

1. **enricher** — adds structure to raw data (Phase 2)
2. **llm-gateway** — required by enricher for LLM calls (Phase 2)
3. **similarity-service** — requires enriched data to be useful (Phase 3)
4. **scoring-service** — requires enriched + similarity data (Phase 3)
5. **notifier** — requires scores to be meaningful (Phase 4)
