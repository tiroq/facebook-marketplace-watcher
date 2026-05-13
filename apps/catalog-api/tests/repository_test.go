package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tiroq/fb-market-watcher/catalog-api/internal/repository"
)

func TestUpsertObservation_NewListing(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "1" {
		t.Skip("skipping integration test: set INTEGRATION_TESTS=1")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required for integration tests")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer pool.Close()

	repo := repository.New(pool)

	price := 28000.0
	result, err := repo.UpsertObservation(ctx, repository.UpsertObservationInput{
		Source:          "facebook_marketplace",
		SourceListingID: "test_" + time.Now().Format("20060102150405"),
		Title:           "Test Listing",
		PriceAmount:     &price,
		Currency:        "THB",
		ObservedAt:      time.Now(),
		RawJSON:         map[string]interface{}{},
	})

	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}
	if result.ListingID == "" {
		t.Error("expected listing ID to be set")
	}
	if !result.Created {
		t.Error("expected listing to be created")
	}
}

func TestUpsertObservation_ExistingListing(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "1" {
		t.Skip("skipping integration test: set INTEGRATION_TESTS=1")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required for integration tests")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer pool.Close()

	repo := repository.New(pool)

	sourceListingID := "existing_" + time.Now().Format("20060102150405")

	price1 := 28000.0
	_, err = repo.UpsertObservation(ctx, repository.UpsertObservationInput{
		Source:          "facebook_marketplace",
		SourceListingID: sourceListingID,
		Title:           "Original Title",
		PriceAmount:     &price1,
		Currency:        "THB",
		ObservedAt:      time.Now(),
		RawJSON:         map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("first upsert failed: %v", err)
	}

	price2 := 25000.0
	result, err := repo.UpsertObservation(ctx, repository.UpsertObservationInput{
		Source:          "facebook_marketplace",
		SourceListingID: sourceListingID,
		Title:           "Original Title",
		PriceAmount:     &price2,
		Currency:        "THB",
		ObservedAt:      time.Now(),
		RawJSON:         map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("second upsert failed: %v", err)
	}
	if result.Created {
		t.Error("expected listing to already exist")
	}
	if !result.Changed {
		t.Error("expected listing to be changed (price changed)")
	}
}

func TestHTTPHealthz(t *testing.T) {
	t.Log("HTTP healthz handler would be tested here with httptest")
}
