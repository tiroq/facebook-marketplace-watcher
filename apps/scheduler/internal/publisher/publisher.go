package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/tiroq/fb-market-watcher/internal/contracts/events"
	"github.com/tiroq/fb-market-watcher/internal/contracts/subjects"
)

// Publisher handles publishing search task events to NATS.
type Publisher struct {
	js     jetstream.JetStream
	dryRun bool
}

// New creates a new Publisher.
func New(js jetstream.JetStream, dryRun bool) *Publisher {
	return &Publisher{js: js, dryRun: dryRun}
}

// PublishSearchRequested publishes a SearchRequested event to NATS.
func (p *Publisher) PublishSearchRequested(ctx context.Context, query, locationHint string, maxCards, maxScrolls int) error {
	eventID := uuid.New().String()
	correlationID := uuid.New().String()

	data := events.SearchRequestedData{
		Query:        query,
		Source:       "facebook_marketplace",
		LocationHint: locationHint,
		MaxCards:     maxCards,
		MaxScrolls:   maxScrolls,
		RequestedAt:  time.Now().UTC(),
		Priority:     1,
	}

	envelope := events.NewEnvelope(eventID, "search.requested", "scheduler", correlationID, data)

	payload, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if p.dryRun {
		slog.Info("DRY RUN: would publish search.requested",
			"query", query,
			"location_hint", locationHint,
			"max_cards", maxCards,
			"event_id", eventID,
		)
		return nil
	}

	ack, err := p.js.Publish(ctx, subjects.SearchRequested, payload)
	if err != nil {
		return fmt.Errorf("publish search.requested: %w", err)
	}

	slog.Info("published search.requested",
		"event_id", eventID,
		"query", query,
		"seq", ack.Sequence,
	)

	return nil
}
