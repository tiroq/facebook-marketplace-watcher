# Events

## Envelope

All NATS events use a standard envelope:

```json
{
  "schema": "fb_event",
  "schema_version": 1,
  "event_id": "550e8400-e29b-41d4-a716-446655440000",
  "event_type": "search.requested",
  "source_service": "scheduler",
  "occurred_at": "2026-05-13T10:00:00Z",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440001",
  "causation_id": "550e8400-e29b-41d4-a716-446655440002",
  "data": {}
}
```

## Subjects

### Implemented in MVP

| Subject | Publisher | Consumer |
|---------|-----------|----------|
| `fb.search.requested` | scheduler | grabber |
| `fb.search.started` | grabber | (log only) |
| `fb.search.completed` | grabber | (log only) |
| `fb.search.failed` | grabber | (log only) |
| `fb.listing.card.observed` | grabber | catalog-api |
| `fb.listing.upserted` | catalog-api | (future) |
| `fb.ops.error` | any | (log only) |

### Future Subjects

| Subject | Purpose |
|---------|---------|
| `fb.listing.detail.requested` | Request detail page scraping |
| `fb.listing.detail.observed` | Detail page scraped |
| `fb.listing.publisher.observed` | Publisher profile observed |
| `fb.listing.enrichment.requested` | Trigger enricher |
| `fb.listing.enriched` | Enrichment complete |
| `fb.listing.similarity.requested` | Trigger similarity check |
| `fb.listing.similarity.completed` | Similarity check complete |
| `fb.listing.duplicate_candidate.created` | Duplicate found |
| `fb.listing.score.requested` | Trigger scoring |
| `fb.listing.score.updated` | Score updated |
| `fb.alert.candidate.created` | Alert worthy listing found |
| `fb.ops.health` | Health heartbeat |

## Payload Examples

### SearchRequested

```json
{
  "query": "macbook pro",
  "source": "facebook_marketplace",
  "location_hint": "Bangkok",
  "max_cards": 50,
  "max_scrolls": 5,
  "requested_at": "2026-05-13T10:00:00Z",
  "priority": 1
}
```

### ListingCardObserved

```json
{
  "search_session_id": "550e8400-e29b-41d4-a716-446655440000",
  "source": "facebook_marketplace",
  "source_listing_id": "123456789",
  "observed_at": "2026-05-13T10:01:00Z",
  "position": 1,
  "url": "https://www.facebook.com/marketplace/item/123456789/",
  "title": "MacBook Pro 15 32GB 1TB",
  "price_text": "฿28,000",
  "price_amount": 28000,
  "currency": "THB",
  "location_text": "Bangkok",
  "raw_json": {}
}
```

### ListingUpserted

```json
{
  "listing_id": "550e8400-e29b-41d4-a716-446655440000",
  "source": "facebook_marketplace",
  "source_listing_id": "123456789",
  "created": true,
  "changed": true,
  "observed_at": "2026-05-13T10:01:00Z"
}
```

## Stream Configuration

- Name: `FB_EVENTS`
- Subjects: `fb.>`
- Storage: File
- Retention: Limits
- Created by: first service to start (idempotent)
