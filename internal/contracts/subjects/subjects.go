package subjects

const (
	// Search lifecycle subjects
	SearchRequested = "fb.search.requested"
	SearchStarted   = "fb.search.started"
	SearchCompleted = "fb.search.completed"
	SearchFailed    = "fb.search.failed"

	// Listing subjects
	ListingCardObserved              = "fb.listing.card.observed"
	ListingDetailRequested           = "fb.listing.detail.requested"
	ListingDetailObserved            = "fb.listing.detail.observed"
	ListingUpserted                  = "fb.listing.upserted"
	ListingPublisherObserved         = "fb.listing.publisher.observed"
	ListingEnrichmentRequested       = "fb.listing.enrichment.requested"
	ListingEnriched                  = "fb.listing.enriched"
	ListingSimilarityRequested       = "fb.listing.similarity.requested"
	ListingSimilarityCompleted       = "fb.listing.similarity.completed"
	ListingDuplicateCandidateCreated = "fb.listing.duplicate_candidate.created"
	ListingScoreRequested            = "fb.listing.score.requested"
	ListingScoreUpdated              = "fb.listing.score.updated"

	// Alert subjects
	AlertCandidateCreated = "fb.alert.candidate.created"

	// Ops subjects
	OpsError  = "fb.ops.error"
	OpsHealth = "fb.ops.health"

	// Stream configuration
	StreamName     = "FB_EVENTS"
	StreamSubjects = "fb.>"
)
