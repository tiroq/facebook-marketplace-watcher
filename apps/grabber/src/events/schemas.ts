import { z } from 'zod';

// Envelope schema for validating incoming NATS events
export const EnvelopeSchema = z.object({
  schema: z.string(),
  schema_version: z.number(),
  event_id: z.string(),
  event_type: z.string(),
  source_service: z.string(),
  occurred_at: z.string(),
  correlation_id: z.string(),
  causation_id: z.string().optional(),
  data: z.unknown(),
});

export type Envelope = z.infer<typeof EnvelopeSchema>;

// SearchRequested payload schema
export const SearchRequestedDataSchema = z.object({
  search_session_id: z.string().optional(),
  search_query_id: z.string().optional(),
  query: z.string(),
  source: z.string().default('facebook_marketplace'),
  location_hint: z.string().optional(),
  max_cards: z.number(),
  max_scrolls: z.number(),
  requested_at: z.string(),
  priority: z.number().default(1),
});

export type SearchRequestedData = z.infer<typeof SearchRequestedDataSchema>;

// ListingCardObserved payload
export const ListingCardObservedDataSchema = z.object({
  search_session_id: z.string(),
  search_query_id: z.string().optional(),
  source: z.string(),
  source_listing_id: z.string(),
  source_publisher_id: z.string().optional(),
  publisher_display_name: z.string().optional(),
  observed_at: z.string(),
  position: z.number(),
  url: z.string(),
  canonical_url: z.string().optional(),
  title: z.string(),
  price_text: z.string().optional(),
  price_amount: z.number().optional(),
  currency: z.string().optional(),
  location_text: z.string().optional(),
  raw_text: z.string().optional(),
  raw_json: z.record(z.unknown()),
  screenshot_path: z.string().optional(),
  html_hash: z.string().optional(),
});

export type ListingCardObservedData = z.infer<typeof ListingCardObservedDataSchema>;
