# Deploy / Compose

This directory is reserved for production or staging variants of the Docker Compose setup.

## Files

| File | Purpose |
|------|---------|
| (root) `docker-compose.yml` | Development all-in-one stack |
| `deploy/compose/` | Production-tuned overrides (e.g. resource limits, external secrets) |

## Running the dev stack

```bash
cp .env.example .env
# Edit .env with your values
task up
task logs
```

## Customising for production

Create a `docker-compose.override.yml` alongside the root `docker-compose.yml` and add
resource constraints, external network configuration, or secret injection as needed.
Docker Compose merges override files automatically.
