package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/tiroq/fb-market-watcher/catalog-api/internal/repository"
	"github.com/tiroq/fb-market-watcher/internal/contracts/events"
	"github.com/tiroq/fb-market-watcher/internal/contracts/subjects"
)

// Service contains the business logic for the catalog API.
type Service struct {
	repo *repository.Repository
	js   jetstream.JetStream
}

// New creates a new Service.
func New(repo *repository.Repository, js jetstream.JetStream) *Service {
	return &Service{repo: repo, js: js}
}

// ObservationRequest is the input for processing a listing card observation.
type ObservationRequest struct {
	SearchSessionID      string                 `json:"search_session_id"`
	SearchQueryID        string                 `json:"search_query_id"`
	Source               string                 `json:"source"`
	SourceListingID      string                 `json:"source_listing_id"`
	SourcePublisherID    string                 `json:"source_publisher_id"`
	PublisherDisplayName string                 `json:"publisher_display_name"`
	ObservedAt           time.Time              `json:"observed_at"`
	Position             int                    `json:"position"`
	URL                  string                 `json:"url"`
	CanonicalURL         string                 `json:"canonical_url"`
	Title                string                 `json:"title"`
	PriceText            string                 `json:"price_text"`
	PriceAmount          *float64               `json:"price_amount"`
	Currency             string                 `json:"currency"`
	LocationText         string                 `json:"location_text"`
	RawText              string                 `json:"raw_text"`
	RawJSON              map[string]interface{} `json:"raw_json"`
	ScreenshotPath       string                 `json:"screenshot_path"`
	HTMLHash             string                 `json:"html_hash"`
}

// ObservationResponse is the response after processing a listing card observation.
type ObservationResponse struct {
	ListingID     string `json:"listing_id"`
	PublisherID   string `json:"publisher_id"`
	Created       bool   `json:"created"`
	Changed       bool   `json:"changed"`
	ObservationID string `json:"observation_id"`
}

// HandleObservation processes a listing card observation, upserts data, and publishes events.
func (s *Service) HandleObservation(ctx context.Context, req ObservationRequest) (ObservationResponse, error) {
	if req.Source == "" {
		req.Source = "facebook_marketplace"
	}
	if req.ObservedAt.IsZero() {
		req.ObservedAt = time.Now().UTC()
	}
	if req.RawJSON == nil {
		req.RawJSON = map[string]interface{}{}
	}

	if req.SourceListingID == "" {
		return ObservationResponse{}, fmt.Errorf("source_listing_id is required")
	}
	if req.Title == "" {
		return ObservationResponse{}, fmt.Errorf("title is required")
	}

	result, err := s.repo.UpsertObservation(ctx, repository.UpsertObservationInput{
		SearchSessionID:      req.SearchSessionID,
		SearchQueryID:        req.SearchQueryID,
		Source:               req.Source,
		SourceListingID:      req.SourceListingID,
		SourcePublisherID:    req.SourcePublisherID,
		PublisherDisplayName: req.PublisherDisplayName,
		ObservedAt:           req.ObservedAt,
		Position:             req.Position,
		URL:                  req.URL,
		CanonicalURL:         req.CanonicalURL,
		Title:                req.Title,
		PriceText:            req.PriceText,
		PriceAmount:          req.PriceAmount,
		Currency:             req.Currency,
		LocationText:         req.LocationText,
		RawText:              req.RawText,
		RawJSON:              req.RawJSON,
		ScreenshotPath:       req.ScreenshotPath,
		HTMLHash:             req.HTMLHash,
	})
	if err != nil {
		return ObservationResponse{}, fmt.Errorf("upsert observation: %w", err)
	}

	slog.Info("listing upserted",
		"listing_id", result.ListingID,
		"source_listing_id", req.SourceListingID,
		"created", result.Created,
		"changed", result.Changed,
	)

	eventID := uuid.New().String()
	envelope := events.NewEnvelope(eventID, "listing.upserted", "catalog-api", req.SearchSessionID, events.ListingUpsertedData{
		ListingID:       result.ListingID,
		Source:          req.Source,
		SourceListingID: req.SourceListingID,
		PublisherID:     result.PublisherID,
		Created:         result.Created,
		Changed:         result.Changed,
		ObservedAt:      req.ObservedAt,
	})

	payload, err := json.Marshal(envelope)
	if err != nil {
		slog.Error("failed to marshal listing.upserted event", "error", err)
	} else {
		if _, err := s.js.Publish(ctx, subjects.ListingUpserted, payload); err != nil {
			slog.Error("failed to publish listing.upserted event", "error", err)
		}
	}

	return ObservationResponse{
		ListingID:     result.ListingID,
		PublisherID:   result.PublisherID,
		Created:       result.Created,
		Changed:       result.Changed,
		ObservationID: result.ObservationID,
	}, nil
}
