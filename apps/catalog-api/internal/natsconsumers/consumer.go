package natsconsumers

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/tiroq/fb-market-watcher/catalog-api/internal/service"
	"github.com/tiroq/fb-market-watcher/internal/contracts/events"
	"github.com/tiroq/fb-market-watcher/internal/contracts/subjects"
)

// Consumer handles NATS JetStream message consumption.
type Consumer struct {
	js  jetstream.JetStream
	svc *service.Service
}

// New creates a new Consumer.
func New(js jetstream.JetStream, svc *service.Service) *Consumer {
	return &Consumer{js: js, svc: svc}
}

// Start subscribes to relevant NATS subjects and processes messages.
func (c *Consumer) Start(ctx context.Context) error {
	cons, err := c.js.CreateOrUpdateConsumer(ctx, subjects.StreamName, jetstream.ConsumerConfig{
		Name:          "catalog-api-card-observer",
		Durable:       "catalog-api-card-observer",
		FilterSubject: subjects.ListingCardObserved,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return err
	}

	cc, err := cons.Consume(func(msg jetstream.Msg) {
		if err := c.handleListingCardObserved(ctx, msg); err != nil {
			slog.Error("failed to handle listing card observed", "error", err)
			msg.Nak()
			return
		}
		msg.Ack()
	})
	if err != nil {
		return err
	}

	slog.Info("NATS consumer started", "subject", subjects.ListingCardObserved)

	<-ctx.Done()
	cc.Stop()
	return nil
}

func (c *Consumer) handleListingCardObserved(ctx context.Context, msg jetstream.Msg) error {
	var envelope events.Envelope
	if err := json.Unmarshal(msg.Data(), &envelope); err != nil {
		slog.Error("failed to unmarshal envelope", "error", err)
		return nil // Don't retry malformed messages
	}

	dataBytes, err := json.Marshal(envelope.Data)
	if err != nil {
		return nil
	}

	var data events.ListingCardObservedData
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		slog.Error("failed to unmarshal listing card observed data", "error", err)
		return nil
	}

	req := service.ObservationRequest{
		SearchSessionID:      data.SearchSessionID,
		SearchQueryID:        data.SearchQueryID,
		Source:               data.Source,
		SourceListingID:      data.SourceListingID,
		SourcePublisherID:    data.SourcePublisherID,
		PublisherDisplayName: data.PublisherDisplayName,
		ObservedAt:           data.ObservedAt,
		Position:             data.Position,
		URL:                  data.URL,
		CanonicalURL:         data.CanonicalURL,
		Title:                data.Title,
		PriceText:            data.PriceText,
		PriceAmount:          data.PriceAmount,
		Currency:             data.Currency,
		LocationText:         data.LocationText,
		RawText:              data.RawText,
		RawJSON:              data.RawJSON,
		ScreenshotPath:       data.ScreenshotPath,
		HTMLHash:             data.HTMLHash,
	}

	_, err = c.svc.HandleObservation(ctx, req)
	if err != nil {
		slog.Error("failed to handle observation from NATS", "error", err, "subject", msg.Subject())
		return err
	}

	return nil
}
