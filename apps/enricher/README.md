# enricher

> **Status: Not implemented in MVP.**

The enricher transforms raw listing observations into structured, normalized metadata.

## Responsibilities

- Text cleanup (remove noise, normalize whitespace)
- Language detection
- Translation to English (via llm-gateway)
- Category and subcategory classification
- Manufacturer and model extraction
- Specification extraction (RAM, storage, screen size, etc.)
- Confidence scoring for each extracted field
- Writes to `listing_metadata` table

## Interface

Consumes: `fb.listing.upserted` (when enrichment_status = not_requested)
Produces: `fb.listing.enrichment.requested` → llm-gateway → `fb.listing.enriched`
Updates: `listing_metadata`, `listings.enrichment_status`

## Why not in MVP

Enrichment requires LLM integration. The raw data in `listings` and
`listing_observations` is sufficient for MVP browsing and filtering.

## Suggested implementation order

1. Implement text cleanup (no LLM needed)
2. Add language detection (langdetect library)
3. Add category classification via simple keyword rules
4. Add LLM-based extraction once llm-gateway is ready
