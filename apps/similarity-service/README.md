# similarity-service

> **Status: Not implemented in MVP.**

The similarity service detects duplicate listings and groups similar products
into product clusters.

## Responsibilities

- Image deduplication via SHA-256 hash
- Perceptual hash (pHash) for near-duplicate images
- Text embedding comparison for similar titles
- Duplicate candidate creation in `duplicate_candidates` table
- Product cluster management in `product_clusters` table
- Prefer local/free algorithms first (no paid API needed)

## Interface

Consumes: `fb.listing.similarity.requested`
Produces: `fb.listing.similarity.completed`, `fb.listing.duplicate_candidate.created`
Updates: `duplicate_candidates`, `product_clusters`, `listings.duplicate_status`

## Why not in MVP

Similarity detection is computationally expensive and requires a corpus of
listings to be useful. Implement after the data collection pipeline is stable.

## Suggested implementation order

1. SHA-256 dedup for images (trivial, high value)
2. Title fuzzy match (Levenshtein distance)
3. pHash for images
4. Vector embeddings (use local model like sentence-transformers)
5. Product cluster assignment
