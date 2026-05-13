import { describe, it, expect } from '@jest/globals';
import { EnvelopeSchema, SearchRequestedDataSchema } from './schemas.js';

describe('EnvelopeSchema', () => {
  it('validates a valid envelope', () => {
    const envelope = {
      schema: 'fb_event',
      schema_version: 1,
      event_id: '550e8400-e29b-41d4-a716-446655440000',
      event_type: 'search.requested',
      source_service: 'scheduler',
      occurred_at: '2026-05-13T10:00:00Z',
      correlation_id: '550e8400-e29b-41d4-a716-446655440001',
      data: {},
    };

    const result = EnvelopeSchema.safeParse(envelope);
    expect(result.success).toBe(true);
  });

  it('rejects envelope without event_id', () => {
    const envelope = {
      schema: 'fb_event',
      schema_version: 1,
      event_type: 'search.requested',
      source_service: 'scheduler',
      occurred_at: '2026-05-13T10:00:00Z',
      correlation_id: '550e8400-e29b-41d4-a716-446655440001',
      data: {},
    };

    const result = EnvelopeSchema.safeParse(envelope);
    expect(result.success).toBe(false);
  });
});

describe('SearchRequestedDataSchema', () => {
  it('validates a valid search requested payload', () => {
    const data = {
      query: 'laptop',
      source: 'facebook_marketplace',
      max_cards: 50,
      max_scrolls: 5,
      requested_at: '2026-05-13T10:00:00Z',
    };

    const result = SearchRequestedDataSchema.safeParse(data);
    expect(result.success).toBe(true);
  });

  it('rejects payload without query', () => {
    const data = {
      source: 'facebook_marketplace',
      max_cards: 50,
      max_scrolls: 5,
      requested_at: '2026-05-13T10:00:00Z',
    };

    const result = SearchRequestedDataSchema.safeParse(data);
    expect(result.success).toBe(false);
  });
});
