package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UpsertObservationInput is the data needed to upsert a listing observation.
type UpsertObservationInput struct {
	SearchSessionID      string
	SearchQueryID        string
	Source               string
	SourceListingID      string
	SourcePublisherID    string
	PublisherDisplayName string
	ObservedAt           time.Time
	Position             int
	URL                  string
	CanonicalURL         string
	Title                string
	PriceText            string
	PriceAmount          *float64
	Currency             string
	LocationText         string
	RawText              string
	RawJSON              map[string]interface{}
	ScreenshotPath       string
	HTMLHash             string
}

// UpsertObservationResult is the result of upserting a listing observation.
type UpsertObservationResult struct {
	ListingID     string
	PublisherID   string
	ObservationID string
	Created       bool
	Changed       bool
}

// Repository handles database operations.
type Repository struct {
	pool *pgxpool.Pool
}

// New creates a new Repository.
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// UpsertObservation atomically upserts publisher, listing, listing_observation, and price_observation.
func (r *Repository) UpsertObservation(ctx context.Context, input UpsertObservationInput) (UpsertObservationResult, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return UpsertObservationResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var publisherID string

	if input.SourcePublisherID != "" {
		publisherID, err = upsertPublisher(ctx, tx, input)
		if err != nil {
			return UpsertObservationResult{}, fmt.Errorf("upsert publisher: %w", err)
		}
	}

	result, err := upsertListing(ctx, tx, input, publisherID)
	if err != nil {
		return UpsertObservationResult{}, fmt.Errorf("upsert listing: %w", err)
	}

	observationID, err := insertListingObservation(ctx, tx, input, result.ListingID)
	if err != nil {
		return UpsertObservationResult{}, fmt.Errorf("insert listing observation: %w", err)
	}
	result.ObservationID = observationID
	result.PublisherID = publisherID

	if input.PriceAmount != nil {
		if err := insertPriceObservation(ctx, tx, input, result.ListingID); err != nil {
			return UpsertObservationResult{}, fmt.Errorf("insert price observation: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return UpsertObservationResult{}, fmt.Errorf("commit tx: %w", err)
	}

	return result, nil
}

func upsertPublisher(ctx context.Context, tx pgx.Tx, input UpsertObservationInput) (string, error) {
	var id string
	err := tx.QueryRow(ctx, `
		INSERT INTO publishers (source, source_publisher_id, display_name, last_seen_at, raw_json)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (source, source_publisher_id)
		WHERE source_publisher_id IS NOT NULL
		DO UPDATE SET
			display_name = COALESCE(EXCLUDED.display_name, publishers.display_name),
			last_seen_at = EXCLUDED.last_seen_at,
			updated_at = now()
		RETURNING id
	`, input.Source, input.SourcePublisherID, input.PublisherDisplayName, input.ObservedAt, "{}").Scan(&id)
	return id, err
}

func upsertListing(ctx context.Context, tx pgx.Tx, input UpsertObservationInput, publisherID string) (UpsertObservationResult, error) {
	rawJSON, err := json.Marshal(input.RawJSON)
	if err != nil {
		rawJSON = []byte("{}")
	}

	canonicalURLHash := ""
	if input.CanonicalURL != "" {
		h := sha256.Sum256([]byte(input.CanonicalURL))
		canonicalURLHash = hex.EncodeToString(h[:])
	}

	var result UpsertObservationResult
	var publisherIDParam interface{} = nil
	if publisherID != "" {
		publisherIDParam = publisherID
	}

	var existingID string
	err = tx.QueryRow(ctx, `
		SELECT id FROM listings WHERE source = $1 AND source_listing_id = $2
	`, input.Source, input.SourceListingID).Scan(&existingID)

	if err == pgx.ErrNoRows {
		err = tx.QueryRow(ctx, `
			INSERT INTO listings (
				source, source_listing_id, publisher_id, canonical_url, canonical_url_hash,
				title_current, price_current, currency, location_current,
				first_seen_at, last_seen_at, raw_json
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10, $11)
			RETURNING id
		`, input.Source, input.SourceListingID, publisherIDParam,
			input.CanonicalURL, canonicalURLHash,
			input.Title, input.PriceAmount, input.Currency, input.LocationText,
			input.ObservedAt, rawJSON).Scan(&result.ListingID)
		if err != nil {
			return result, err
		}
		result.Created = true
		result.Changed = true
	} else if err != nil {
		return result, err
	} else {
		result.ListingID = existingID
		var oldTitle string
		var oldPrice *float64
		var oldLocation *string

		err = tx.QueryRow(ctx, `
			SELECT title_current, price_current, location_current FROM listings WHERE id = $1
		`, existingID).Scan(&oldTitle, &oldPrice, &oldLocation)
		if err != nil {
			return result, err
		}

		changed := oldTitle != input.Title ||
			(oldPrice == nil && input.PriceAmount != nil) ||
			(oldPrice != nil && input.PriceAmount == nil) ||
			(oldPrice != nil && input.PriceAmount != nil && *oldPrice != *input.PriceAmount) ||
			(oldLocation == nil && input.LocationText != "") ||
			(oldLocation != nil && *oldLocation != input.LocationText)

		result.Changed = changed

		if changed {
			_, err = tx.Exec(ctx, `
				UPDATE listings SET
					publisher_id = COALESCE($1, publisher_id),
					canonical_url = COALESCE($2, canonical_url),
					canonical_url_hash = COALESCE($3, canonical_url_hash),
					title_current = $4,
					price_current = $5,
					currency = COALESCE($6, currency),
					location_current = $7,
					last_seen_at = $8,
					updated_at = now(),
					version = version + 1
				WHERE id = $9
			`,
				publisherIDParam, input.CanonicalURL, canonicalURLHash,
				input.Title, input.PriceAmount, input.Currency, input.LocationText,
				input.ObservedAt, existingID)
		} else {
			_, err = tx.Exec(ctx, `
				UPDATE listings SET
					publisher_id = COALESCE($1, publisher_id),
					canonical_url = COALESCE($2, canonical_url),
					canonical_url_hash = COALESCE($3, canonical_url_hash),
					title_current = $4,
					price_current = $5,
					currency = COALESCE($6, currency),
					location_current = $7,
					last_seen_at = $8,
					updated_at = now()
				WHERE id = $9
			`,
				publisherIDParam, input.CanonicalURL, canonicalURLHash,
				input.Title, input.PriceAmount, input.Currency, input.LocationText,
				input.ObservedAt, existingID)
		}
		if err != nil {
			return result, err
		}
	}

	return result, nil
}

func insertListingObservation(ctx context.Context, tx pgx.Tx, input UpsertObservationInput, listingID string) (string, error) {
	rawJSON, _ := json.Marshal(input.RawJSON)

	id := uuid.New().String()
	var searchSessionIDParam interface{} = nil
	if input.SearchSessionID != "" {
		searchSessionIDParam = input.SearchSessionID
	}
	var searchQueryIDParam interface{} = nil
	if input.SearchQueryID != "" {
		searchQueryIDParam = input.SearchQueryID
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO listing_observations (
			id, listing_id, search_session_id, search_query_id, observed_at, position,
			title_observed, price_observed, price_text, currency, location_observed,
			url_observed, raw_text, raw_json, screenshot_path, html_hash
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, id, listingID, searchSessionIDParam, searchQueryIDParam, input.ObservedAt, input.Position,
		input.Title, input.PriceAmount, input.PriceText, input.Currency, input.LocationText,
		input.URL, input.RawText, rawJSON, input.ScreenshotPath, input.HTMLHash)

	return id, err
}

func insertPriceObservation(ctx context.Context, tx pgx.Tx, input UpsertObservationInput, listingID string) error {
	var searchQueryIDParam interface{} = nil
	if input.SearchQueryID != "" {
		searchQueryIDParam = input.SearchQueryID
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO price_observations (listing_id, observed_at, price_amount, currency, source, search_query_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, listingID, input.ObservedAt, input.PriceAmount, input.Currency, input.Source, searchQueryIDParam)
	return err
}

// Listing represents a listing record for API responses.
type Listing struct {
	ID              string    `json:"id"`
	Source          string    `json:"source"`
	SourceListingID string    `json:"source_listing_id"`
	PublisherID     *string   `json:"publisher_id,omitempty"`
	CanonicalURL    *string   `json:"canonical_url,omitempty"`
	TitleCurrent    *string   `json:"title_current,omitempty"`
	PriceCurrent    *float64  `json:"price_current,omitempty"`
	Currency        *string   `json:"currency,omitempty"`
	LocationCurrent *string   `json:"location_current,omitempty"`
	Status          string    `json:"status"`
	FirstSeenAt     time.Time `json:"first_seen_at"`
	LastSeenAt      time.Time `json:"last_seen_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Version         int64     `json:"version"`
}

// ListListingsFilter defines query filters for listing search.
type ListListingsFilter struct {
	Limit  int
	Offset int
	Source string
	Status string
	Query  string
}

// ListListings returns a paginated list of listings.
func (r *Repository) ListListings(ctx context.Context, f ListListingsFilter) ([]Listing, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}

	query := `
		SELECT id, source, source_listing_id, publisher_id, canonical_url,
		       title_current, price_current, currency, location_current,
		       status, first_seen_at, last_seen_at, created_at, updated_at, version
		FROM listings
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if f.Source != "" {
		query += fmt.Sprintf(" AND source = $%d", argIdx)
		args = append(args, f.Source)
		argIdx++
	}
	if f.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, f.Status)
		argIdx++
	}
	if f.Query != "" && len(f.Query) <= 200 {
		query += fmt.Sprintf(" AND title_current ILIKE $%d", argIdx)
		args = append(args, "%"+f.Query+"%")
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY last_seen_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []Listing
	for rows.Next() {
		var l Listing
		err := rows.Scan(
			&l.ID, &l.Source, &l.SourceListingID, &l.PublisherID, &l.CanonicalURL,
			&l.TitleCurrent, &l.PriceCurrent, &l.Currency, &l.LocationCurrent,
			&l.Status, &l.FirstSeenAt, &l.LastSeenAt, &l.CreatedAt, &l.UpdatedAt, &l.Version,
		)
		if err != nil {
			return nil, err
		}
		listings = append(listings, l)
	}

	return listings, rows.Err()
}

// GetListing returns a single listing by ID.
func (r *Repository) GetListing(ctx context.Context, id string) (*Listing, error) {
	var l Listing
	err := r.pool.QueryRow(ctx, `
		SELECT id, source, source_listing_id, publisher_id, canonical_url,
		       title_current, price_current, currency, location_current,
		       status, first_seen_at, last_seen_at, created_at, updated_at, version
		FROM listings WHERE id = $1
	`, id).Scan(
		&l.ID, &l.Source, &l.SourceListingID, &l.PublisherID, &l.CanonicalURL,
		&l.TitleCurrent, &l.PriceCurrent, &l.Currency, &l.LocationCurrent,
		&l.Status, &l.FirstSeenAt, &l.LastSeenAt, &l.CreatedAt, &l.UpdatedAt, &l.Version,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// SearchQuery represents a search_queries record.
type SearchQuery struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Query            string    `json:"query"`
	Source           string    `json:"source"`
	LocationHint     *string   `json:"location_hint,omitempty"`
	Enabled          bool      `json:"enabled"`
	MaxCardsPerRun   int       `json:"max_cards_per_run"`
	MaxScrollsPerRun int       `json:"max_scrolls_per_run"`
	Timezone         string    `json:"timezone"`
	IntervalMinutes  int       `json:"interval_minutes"`
	JitterMinutes    int       `json:"jitter_minutes"`
	SkipProbability  float64   `json:"skip_probability"`
	MaxRunsPerDay    int       `json:"max_runs_per_day"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ListSearchQueries returns all search queries.
func (r *Repository) ListSearchQueries(ctx context.Context) ([]SearchQuery, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, query, source, location_hint, enabled,
		       max_cards_per_run, max_scrolls_per_run, timezone,
		       interval_minutes, jitter_minutes, skip_probability, max_runs_per_day,
		       created_at, updated_at
		FROM search_queries ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queries []SearchQuery
	for rows.Next() {
		var q SearchQuery
		err := rows.Scan(
			&q.ID, &q.Name, &q.Query, &q.Source, &q.LocationHint, &q.Enabled,
			&q.MaxCardsPerRun, &q.MaxScrollsPerRun, &q.Timezone,
			&q.IntervalMinutes, &q.JitterMinutes, &q.SkipProbability, &q.MaxRunsPerDay,
			&q.CreatedAt, &q.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		queries = append(queries, q)
	}
	return queries, rows.Err()
}

// CreateSearchQueryInput is the input for creating a search query.
type CreateSearchQueryInput struct {
	Name             string  `json:"name"`
	Query            string  `json:"query"`
	Source           string  `json:"source"`
	LocationHint     string  `json:"location_hint"`
	MaxCardsPerRun   int     `json:"max_cards_per_run"`
	MaxScrollsPerRun int     `json:"max_scrolls_per_run"`
	Timezone         string  `json:"timezone"`
	IntervalMinutes  int     `json:"interval_minutes"`
	JitterMinutes    int     `json:"jitter_minutes"`
	SkipProbability  float64 `json:"skip_probability"`
	MaxRunsPerDay    int     `json:"max_runs_per_day"`
}

// CreateSearchQuery inserts a new search query.
func (r *Repository) CreateSearchQuery(ctx context.Context, input CreateSearchQueryInput) (*SearchQuery, error) {
	if input.Source == "" {
		input.Source = "facebook_marketplace"
	}
	if input.Timezone == "" {
		input.Timezone = "Asia/Bangkok"
	}
	if input.MaxCardsPerRun == 0 {
		input.MaxCardsPerRun = 50
	}
	if input.MaxScrollsPerRun == 0 {
		input.MaxScrollsPerRun = 5
	}
	if input.IntervalMinutes == 0 {
		input.IntervalMinutes = 40
	}
	if input.JitterMinutes == 0 {
		input.JitterMinutes = 10
	}
	if input.SkipProbability == 0 {
		input.SkipProbability = 0.5
	}
	if input.MaxRunsPerDay == 0 {
		input.MaxRunsPerDay = 3
	}

	var locationHintParam interface{} = nil
	if input.LocationHint != "" {
		locationHintParam = input.LocationHint
	}

	var q SearchQuery
	err := r.pool.QueryRow(ctx, `
		INSERT INTO search_queries (name, query, source, location_hint, max_cards_per_run, max_scrolls_per_run,
		                            timezone, interval_minutes, jitter_minutes, skip_probability, max_runs_per_day)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, name, query, source, location_hint, enabled,
		          max_cards_per_run, max_scrolls_per_run, timezone,
		          interval_minutes, jitter_minutes, skip_probability, max_runs_per_day,
		          created_at, updated_at
	`, input.Name, input.Query, input.Source, locationHintParam,
		input.MaxCardsPerRun, input.MaxScrollsPerRun, input.Timezone,
		input.IntervalMinutes, input.JitterMinutes, input.SkipProbability, input.MaxRunsPerDay,
	).Scan(
		&q.ID, &q.Name, &q.Query, &q.Source, &q.LocationHint, &q.Enabled,
		&q.MaxCardsPerRun, &q.MaxScrollsPerRun, &q.Timezone,
		&q.IntervalMinutes, &q.JitterMinutes, &q.SkipProbability, &q.MaxRunsPerDay,
		&q.CreatedAt, &q.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &q, nil
}
