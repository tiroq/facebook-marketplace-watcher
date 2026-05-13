# Safety

## Personal Use Only

fb-market-watcher is designed for **personal, low-frequency use** by a single user
monitoring Facebook Marketplace for items of personal interest.

It is not designed for:
- Commercial scraping at scale
- Reselling scraped data
- Automated purchasing or bidding
- Multi-user deployments
- Any use that violates Facebook's Terms of Service

## No Bypass

The system does **not** attempt to bypass Facebook's bot detection, CAPTCHA,
or login requirements. When the grabber encounters a login wall, checkpoint,
or CAPTCHA, it stops immediately and reports the error via `fb.ops.error`.

No headless mode tricks, CAPTCHA-solving services, or fingerprint spoofing
are used or intended.

## Manual Login Required

The grabber uses a **persistent browser profile** stored in `data/browser-profile/`.
You must log in to Facebook manually using a real browser session before
the grabber can operate:

1. Start the stack: `docker compose up -d`
2. The grabber will fail with a login error on first run
3. Open a browser with the persistent profile and log in manually
4. Restart the grabber: `docker compose restart grabber`

The session cookie is then reused for subsequent runs.

## Rate Limiting by Design

The scheduler applies multiple layers of rate limiting to stay well within
reasonable usage patterns:

- `interval_minutes`: Minimum time between runs (default: 40 minutes)
- `jitter_minutes`: Random jitter to avoid predictable patterns (default: ±10 minutes)
- `skip_probability`: Probability of skipping a scheduled slot (default: 0.5)
- `max_runs_per_day`: Hard cap on daily runs per query (default: 3)
- `active_window_start` / `active_window_end`: Only run during certain hours

This means a single search query runs at most 3 times per day, with random
intervals between runs and random skips. This is consistent with a person
manually checking the marketplace a few times a day.

## Data Retention

All scraped data is stored locally in your own PostgreSQL instance.
No data is sent to any third party (except LLM providers if llm-gateway
is enabled in future phases).

## No Credentials in Code

All credentials (database passwords, NATS credentials, API keys) are
passed via environment variables defined in `.env`. The `.env` file is
excluded from git via `.gitignore`. See `.env.example` for the required
variables.

Never commit credentials to source control.
