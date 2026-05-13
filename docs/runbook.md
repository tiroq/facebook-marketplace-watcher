# Runbook

## Start / Stop

### Start all services
```bash
docker compose up -d
```

### Stop all services
```bash
docker compose down
```

### Stop and remove volumes (full reset)
```bash
docker compose down -v
```

### Restart a single service
```bash
docker compose restart catalog-api
docker compose restart scheduler
docker compose restart grabber
```

### View running services
```bash
docker compose ps
```

## Logs

### Tail logs for all services
```bash
docker compose logs -f
```

### Tail logs for a specific service
```bash
docker compose logs -f catalog-api
docker compose logs -f scheduler
docker compose logs -f grabber
```

### View last N lines
```bash
docker compose logs --tail=100 grabber
```

## Database Access

### Open psql shell
```bash
docker compose exec postgres psql -U postgres -d fb_market_watcher
```

### Common queries

List all listings:
```sql
SELECT id, source_listing_id, title_current, price_current, currency, last_seen_at
FROM listings
ORDER BY last_seen_at DESC
LIMIT 20;
```

Count listings by status:
```sql
SELECT status, count(*) FROM listings GROUP BY status;
```

View recent price observations:
```sql
SELECT l.title_current, po.price_amount, po.currency, po.observed_at
FROM price_observations po
JOIN listings l ON l.id = po.listing_id
ORDER BY po.observed_at DESC
LIMIT 20;
```

View active search queries:
```sql
SELECT name, query, enabled, interval_minutes, skip_probability
FROM search_queries
WHERE enabled = true;
```

Check recent observations:
```sql
SELECT lo.observed_at, lo.title_observed, lo.price_observed, lo.currency
FROM listing_observations lo
ORDER BY lo.observed_at DESC
LIMIT 20;
```

## NATS Checks

### Open NATS CLI shell
```bash
docker compose exec nats nats --server nats://localhost:4222 stream list
```

### View stream info
```bash
docker compose exec nats nats --server nats://localhost:4222 stream info FB_EVENTS
```

### View recent messages on a subject
```bash
docker compose exec nats nats --server nats://localhost:4222 sub 'fb.>' --count=5
```

### View consumer info
```bash
docker compose exec nats nats --server nats://localhost:4222 consumer list FB_EVENTS
```

## Admin UIs

- **NocoDB**: http://localhost:8080 — browse and edit listings
- **Metabase**: http://localhost:3000 — analytics dashboards
- **NATS monitoring**: http://localhost:8222 — stream and connection stats

## Common Failures

### grabber: "login required" or captcha detected
The grabber detected a Facebook checkpoint. Manual intervention required:

1. Stop the grabber: `docker compose stop grabber`
2. Open a browser pointing at the persistent profile in `data/browser-profile/`
3. Log in to Facebook manually
4. Start the grabber: `docker compose start grabber`

### catalog-api: "connection refused" to PostgreSQL
PostgreSQL may not be ready yet. Check:
```bash
docker compose logs postgres | tail -20
docker compose ps postgres
```

Wait for the health check to pass, then restart catalog-api:
```bash
docker compose restart catalog-api
```

### scheduler not emitting events
Check that search queries exist and are enabled:
```sql
SELECT name, enabled, interval_minutes FROM search_queries;
```

If no rows exist, insert a search query via the API:
```bash
curl -X POST http://localhost:8090/v1/search-queries \
  -H 'Content-Type: application/json' \
  -d '{"name":"MacBook Pro","query":"macbook pro","location_hint":"Bangkok"}'
```

### NATS stream not found
The stream is created by the first service that starts. If it is missing:
```bash
docker compose restart catalog-api
```

### High memory usage in grabber
Playwright keeps a browser context open. Restart to free memory:
```bash
docker compose restart grabber
```

## Health Checks

```bash
# catalog-api health
curl http://localhost:8090/healthz

# catalog-api readiness (checks DB connection)
curl http://localhost:8090/readyz
```

Expected response: `{"status":"ok"}`
