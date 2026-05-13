# scoring-service

> **Status: Not implemented in MVP.**

The scoring service evaluates each listing and computes scores to help identify
good deals, risky sellers, and price outliers.

## Responsibilities

- Compute `deal_score`: composite of price, quality, and risk
- Compute `price_score`: how current price compares to market baseline
- Compute `quality_score`: based on listing completeness and image quality
- Compute `risk_score`: based on publisher history and listing patterns
- Compute `freshness_score`: how recently the listing was observed
- Compute `fit_score`: how well the listing matches user preferences
- Compute `fair_price_estimate`: derived from `market_price_baselines`
- Compute `discount_ratio`: (fair_price - current_price) / fair_price
- Use deterministic scoring first; ML scoring later

## Interface

Consumes: `fb.listing.score.requested`
Produces: `fb.listing.score.updated`
Updates: `listing_scores`

## Why not in MVP

Scoring requires a corpus of listings and market baselines to be meaningful.
Implement after enricher and similarity-service are working.

## Suggested implementation order

1. Freshness score (trivial: time since first_seen_at)
2. Price score (requires market_price_baselines to be populated)
3. Risk score (basic rules: new publisher, very low price)
4. Deal score (combine price + quality + risk)
5. ML-based scoring (future)
