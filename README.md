# fb-market-watcher

A personal Facebook Marketplace intelligence system that watches listings, stores
them, enriches with AI metadata, and surfaces deals through a self-hosted dashboard.

> **Personal use only.** This tool does not bypass any security mechanism.
> It uses a real, logged-in browser session that you control.
> Respect Facebook's Terms of Service and your local laws.

---

## Architecture

```
┌─────────────┐   NATS subject          ┌──────────────┐
│  scheduler  │──▶ marketplace.scrape ──▶│   grabber    │
│  (Go)       │   (trigger job)          │  (Node/PW)   │
└─────────────┘                          └──────┬───────┘
                                                │ NATS subject
                                                │ marketplace.card.raw
                                                ▼
                                         ┌──────────────┐
                                         │  catalog-api │
                                         │  (Go / HTTP) │
                                         └──────┬───────┘
                                                │ PostgreSQL
                                                ▼
                               ┌────────────────────────────┐
                               │         postgres:16         │
                               └──────┬──────────┬──────────┘
                                      │          │
                               ┌──────▼──┐  ┌───▼──────┐
                               │ NocoDB  │  │ Metabase │
                               │ :8088   │  │  :3000   │
                               └─────────┘  └──────────┘

Planned services (stubs present):
  enricher · similarity-service · scoring-service · notifier · llm-gateway
```

---

## Quick Start

### 1. Prerequisites

- Docker ≥ 24 with the Compose plugin
- [Task](https://taskfile.dev) (`brew install go-task` or `go install github.com/go-task/task/v3/cmd/task@latest`)

### 2. Configure

```bash
cp .env.example .env
# Open .env and set at minimum:
#   POSTGRES_PASSWORD   – choose a strong password
#   DEFAULT_SEARCH_QUERY, DEFAULT_LOCATION_HINT
```

### 3. Start the stack

```bash
task up
task logs          # follow all logs; Ctrl-C to detach
task ps            # check service health
```

Services will be available at:

| Service     | URL                   |
|-------------|-----------------------|
| catalog-api | http://localhost:8080 |
| NocoDB      | http://localhost:8088 |
| Metabase    | http://localhost:3000 |
| NATS mon.   | http://localhost:8222 |

---

## Preparing the Browser Profile

The grabber uses Playwright with a **persistent browser profile** so you stay
logged in to Facebook between runs.

1. Make sure `data/browser-profile/` exists (it's tracked with `.gitkeep`).
2. Run grabber once with `FB_HEADLESS=false` and manually log in to Facebook.
3. The session is saved in `data/browser-profile/` and reused on every subsequent run.

```bash
# Local (outside Docker) first-time login:
cd apps/grabber
npm install
FB_HEADLESS=false FB_BROWSER_PROFILE_DIR=../../data/browser-profile npm run dev
```

---

## Running Your First Search

Searches are triggered automatically by the **scheduler** on the interval defined
in `.env` (`DEFAULT_INTERVAL_MINUTES`). To trigger one immediately:

```bash
# Via NATS CLI (if installed):
nats pub marketplace.scrape '{"query":"laptop","locationHint":"Bangkok","maxCards":20}'

# Or restart the scheduler with SCHEDULER_DRY_RUN=false and it fires on startup.
```

Results appear in the `cards` table in PostgreSQL within seconds.

---

## Viewing Data

### NocoDB (spreadsheet view)

1. Open http://localhost:8088
2. Create a new project → connect to the existing PostgreSQL (host `postgres`,
   port `5432`, credentials from your `.env`).
3. Browse the `cards` table.

### Metabase (charts & dashboards)

1. Open http://localhost:3000 and complete the first-run wizard.
2. Connect to PostgreSQL: host `postgres`, port `5432`, credentials from `.env`.
3. Build questions and dashboards on the `cards` table.

---

## Dev Commands

```bash
task up                # start stack
task down              # stop stack
task logs              # follow logs
task ps                # service status

task db:migrate        # apply SQL files in sql/migrations/
task db:psql           # open psql shell

task nats:info         # show NATS server stats

task catalog:run       # run catalog-api locally
task scheduler:run     # run scheduler locally

task grabber:install   # npm install for grabber
task grabber:run       # run grabber in dev mode

task test              # run all tests
task test:go           # Go tests only
task test:grabber      # grabber tests only

task fmt               # format all code
task lint              # vet / lint all code

task clean             # remove containers, volumes, build artifacts
```

---

## Safety Notes

- **Personal use only.** Do not deploy this for third parties.
- The grabber uses a real browser and your real Facebook session. It does not
  crack CAPTCHAs or bypass any security measure.
- Use generous intervals (`DEFAULT_INTERVAL_MINUTES ≥ 30`) to avoid triggering
  rate-limits or account restrictions.
- Do not commit your `.env` file or `data/browser-profile/` to version control.

---

## Project Status

| Component          | Status      |
|--------------------|-------------|
| Infrastructure     | ✅ Ready    |
| catalog-api        | 🚧 Scaffold |
| scheduler          | 🚧 Scaffold |
| grabber            | 🚧 Scaffold |
| enricher           | 📋 Planned  |
| similarity-service | 📋 Planned  |
| scoring-service    | 📋 Planned  |
| notifier           | 📋 Planned  |
| llm-gateway        | 📋 Planned  |
