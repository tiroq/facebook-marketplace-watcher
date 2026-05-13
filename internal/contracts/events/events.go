package events

import "time"

// Envelope is the standard wrapper for all NATS events.
type Envelope struct {
	Schema        string      `json:"schema"`
	SchemaVersion int         `json:"schema_version"`
	EventID       string      `json:"event_id"`
	EventType     string      `json:"event_type"`
	SourceService string      `json:"source_service"`
	OccurredAt    time.Time   `json:"occurred_at"`
	CorrelationID string      `json:"correlation_id"`
	CausationID   string      `json:"causation_id,omitempty"`
	Data          interface{} `json:"data"`
}

// NewEnvelope creates a new event envelope with the standard schema.
func NewEnvelope(eventID, eventType, sourceService, correlationID string, data interface{}) Envelope {
	return Envelope{
		Schema:        "fb_event",
		SchemaVersion: 1,
		EventID:       eventID,
		EventType:     eventType,
		SourceService: sourceService,
		OccurredAt:    time.Now().UTC(),
		CorrelationID: correlationID,
		Data:          data,
	}
}

// SearchRequestedData is the payload for fb.search.requested.
type SearchRequestedData struct {
	SearchSessionID string    `json:"search_session_id,omitempty"`
	SearchQueryID   string    `json:"search_query_id,omitempty"`
	Query           string    `json:"query"`
	Source          string    `json:"source"`
	LocationHint    string    `json:"location_hint,omitempty"`
	MaxCards        int       `json:"max_cards"`
	MaxScrolls      int       `json:"max_scrolls"`
	RequestedAt     time.Time `json:"requested_at"`
	Priority        int       `json:"priority"`
}

// SearchStartedData is the payload for fb.search.started.
type SearchStartedData struct {
	SearchSessionID string    `json:"search_session_id"`
	Query           string    `json:"query"`
	StartedAt       time.Time `json:"started_at"`
}

// SearchCompletedData is the payload for fb.search.completed.
type SearchCompletedData struct {
	SearchSessionID string    `json:"search_session_id"`
	Query           string    `json:"query"`
	CardsObserved   int       `json:"cards_observed"`
	CompletedAt     time.Time `json:"completed_at"`
}

// SearchFailedData is the payload for fb.search.failed.
type SearchFailedData struct {
	SearchSessionID string    `json:"search_session_id,omitempty"`
	Query           string    `json:"query,omitempty"`
	ErrorCode       string    `json:"error_code"`
	ErrorMessage    string    `json:"error_message"`
	FailedAt        time.Time `json:"failed_at"`
	ScreenshotPath  string    `json:"screenshot_path,omitempty"`
}

// ListingCardObservedData is the payload for fb.listing.card.observed.
type ListingCardObservedData struct {
	SearchSessionID      string                 `json:"search_session_id"`
	SearchQueryID        string                 `json:"search_query_id,omitempty"`
	Source               string                 `json:"source"`
	SourceListingID      string                 `json:"source_listing_id"`
	SourcePublisherID    string                 `json:"source_publisher_id,omitempty"`
	PublisherDisplayName string                 `json:"publisher_display_name,omitempty"`
	ObservedAt           time.Time              `json:"observed_at"`
	Position             int                    `json:"position"`
	URL                  string                 `json:"url"`
	CanonicalURL         string                 `json:"canonical_url,omitempty"`
	Title                string                 `json:"title"`
	PriceText            string                 `json:"price_text,omitempty"`
	PriceAmount          *float64               `json:"price_amount,omitempty"`
	Currency             string                 `json:"currency,omitempty"`
	LocationText         string                 `json:"location_text,omitempty"`
	RawText              string                 `json:"raw_text,omitempty"`
	RawJSON              map[string]interface{} `json:"raw_json"`
	ScreenshotPath       string                 `json:"screenshot_path,omitempty"`
	HTMLHash             string                 `json:"html_hash,omitempty"`
}

// ListingUpsertedData is the payload for fb.listing.upserted.
type ListingUpsertedData struct {
	ListingID       string    `json:"listing_id"`
	Source          string    `json:"source"`
	SourceListingID string    `json:"source_listing_id"`
	PublisherID     string    `json:"publisher_id,omitempty"`
	Created         bool      `json:"created"`
	Changed         bool      `json:"changed"`
	ObservedAt      time.Time `json:"observed_at"`
}

// OpsErrorData is the payload for fb.ops.error.
type OpsErrorData struct {
	Service      string `json:"service"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Subject      string `json:"subject,omitempty"`
}
