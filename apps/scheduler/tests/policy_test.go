package tests

import (
	"testing"
	"time"

	"github.com/tiroq/fb-market-watcher/scheduler/internal/policy"
)

func TestShouldRun_SkipProbability_Zero(t *testing.T) {
	p := policy.Policy{
		SkipProbability: 0,
	}

	for i := 0; i < 100; i++ {
		if !policy.ShouldRun(time.Now(), p) {
			t.Error("expected ShouldRun to return true with skip probability 0")
		}
	}
}

func TestShouldRun_SkipProbability_One(t *testing.T) {
	p := policy.Policy{
		SkipProbability: 1.0,
	}

	skipped := 0
	for i := 0; i < 100; i++ {
		if !policy.ShouldRun(time.Now(), p) {
			skipped++
		}
	}
	if skipped != 100 {
		t.Errorf("expected all 100 runs to be skipped, got %d skipped", skipped)
	}
}

func TestNextDelay_NoJitter(t *testing.T) {
	delay := policy.NextDelay(40, 0)
	expected := 40 * time.Minute
	if delay != expected {
		t.Errorf("expected %v, got %v", expected, delay)
	}
}

func TestNextDelay_WithJitter(t *testing.T) {
	base := 40
	jitter := 10

	for i := 0; i < 100; i++ {
		delay := policy.NextDelay(base, jitter)
		minExpected := time.Duration(base) * time.Minute
		maxExpected := time.Duration(base+jitter) * time.Minute

		if delay < minExpected {
			t.Errorf("delay %v is less than minimum %v", delay, minExpected)
		}
		if delay > maxExpected {
			t.Errorf("delay %v exceeds maximum %v", delay, maxExpected)
		}
	}
}

func TestActiveWindow_InWindow(t *testing.T) {
	now := time.Date(2026, 1, 1, 14, 0, 0, 0, time.UTC)

	windowStart := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	windowEnd := time.Date(2026, 1, 1, 18, 0, 0, 0, time.UTC)

	p := policy.Policy{
		SkipProbability:   0,
		ActiveWindowStart: &windowStart,
		ActiveWindowEnd:   &windowEnd,
		Timezone:          "UTC",
	}

	if !policy.ShouldRun(now, p) {
		t.Error("expected ShouldRun to return true within active window")
	}
}

func TestActiveWindow_OutsideWindow(t *testing.T) {
	now := time.Date(2026, 1, 1, 3, 0, 0, 0, time.UTC)

	windowStart := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	windowEnd := time.Date(2026, 1, 1, 18, 0, 0, 0, time.UTC)

	p := policy.Policy{
		SkipProbability:   0,
		ActiveWindowStart: &windowStart,
		ActiveWindowEnd:   &windowEnd,
		Timezone:          "UTC",
	}

	if policy.ShouldRun(now, p) {
		t.Error("expected ShouldRun to return false outside active window")
	}
}
