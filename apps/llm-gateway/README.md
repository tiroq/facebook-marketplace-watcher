# llm-gateway

> **Status: Not implemented in MVP.**

The LLM gateway provides a centralized interface for all LLM API calls in the system.

## Responsibilities

- Rate limiting per provider and globally
- Cache by `input_hash` to avoid redundant API calls
- Provider routing (OpenAI, Anthropic, local Ollama)
- Structured JSON response validation
- Retry with exponential backoff
- Audit logging via `llm_requests` table

## Interface

Consumes: `fb.listing.enrichment.requested`
Publishes: `fb.listing.enriched`, `fb.ops.error`

## Why not in MVP

No LLM calls are needed in MVP. The grabber, catalog-api, and scheduler
form a complete data collection pipeline without LLM enrichment.

## Suggested implementation order

1. Add Ollama integration first (free, local)
2. Add OpenAI as fallback
3. Implement cache layer using `llm_requests.input_hash`
4. Add rate limiter middleware
