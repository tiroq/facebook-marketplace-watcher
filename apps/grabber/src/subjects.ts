// NATS subjects — must match the Go contracts in internal/contracts/subjects/subjects.go
export const subjects = {
  searchRequested: 'fb.search.requested',
  searchStarted: 'fb.search.started',
  searchCompleted: 'fb.search.completed',
  searchFailed: 'fb.search.failed',
  listingCardObserved: 'fb.listing.card.observed',
  listingUpserted: 'fb.listing.upserted',
  opsError: 'fb.ops.error',
} as const;

export const STREAM_NAME = 'FB_EVENTS';
