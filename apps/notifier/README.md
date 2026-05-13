# notifier

> **Status: Not implemented in MVP.**

The notifier sends alerts to the user when interesting listings are found.

## Responsibilities

- Send Telegram messages for high-score listings
- Alert on significant price drops (price_score improvement)
- Flag risky deals that need manual review
- Deduplicate alerts (don't re-alert on already-seen listings)
- Respect quiet hours / active window

## Interface

Consumes: `fb.listing.score.updated`, `fb.alert.candidate.created`
Produces: Telegram messages via Bot API
Updates: `task_audit_log` for sent alerts

## Configuration

- `TELEGRAM_BOT_TOKEN`: Bot API token
- `TELEGRAM_CHAT_ID`: Target chat/channel ID
- `NOTIFY_DEAL_SCORE_THRESHOLD`: Minimum deal score to trigger alert (default: 0.7)
- `NOTIFY_PRICE_DROP_THRESHOLD`: Minimum price drop ratio (default: 0.1)

## Why not in MVP

Notifications are only useful once scoring is implemented. The user can
browse listings directly in NocoDB during MVP.

## Suggested implementation order

1. Telegram bot setup
2. High deal score alerts
3. Price drop alerts
4. Risky deal alerts
5. Alert deduplication
