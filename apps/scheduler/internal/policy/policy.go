package policy

import (
	"math/rand"
	"time"
)

// rng is a package-level random source seeded at startup.
// Using a dedicated source makes the seeding intent explicit and keeps the
// package independent of the global rand state.
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// Policy defines the scheduling policy for a search task.
type Policy struct {
	IntervalMinutes   int
	JitterMinutes     int
	SkipProbability   float64
	ActiveWindowStart *time.Time // nil means no restriction
	ActiveWindowEnd   *time.Time // nil means no restriction
	Timezone          string
}

// ShouldRun determines if the scheduler should run based on the current time and policy.
// It checks skip probability and active window.
func ShouldRun(now time.Time, p Policy) bool {
	if p.ActiveWindowStart != nil && p.ActiveWindowEnd != nil {
		if !inActiveWindow(now, *p.ActiveWindowStart, *p.ActiveWindowEnd, p.Timezone) {
			return false
		}
	}

	if p.SkipProbability > 0 {
		if rng.Float64() < p.SkipProbability {
			return false
		}
	}

	return true
}

// NextDelay returns the next delay duration with jitter applied.
// Jitter is randomly added in the range [0, jitterMinutes), so the total
// delay falls in [intervalMinutes, intervalMinutes+jitterMinutes).
func NextDelay(intervalMinutes, jitterMinutes int) time.Duration {
	base := time.Duration(intervalMinutes) * time.Minute
	if jitterMinutes <= 0 {
		return base
	}
	jitter := time.Duration(rng.Intn(jitterMinutes*60)) * time.Second
	return base + jitter
}

// inActiveWindow checks if the given time is within the active window.
func inActiveWindow(now time.Time, windowStart, windowEnd time.Time, timezone string) bool {
	loc := time.UTC
	if timezone != "" {
		if l, err := time.LoadLocation(timezone); err == nil {
			loc = l
		}
	}

	nowLocal := now.In(loc)

	hour, min, sec := nowLocal.Clock()
	nowSeconds := hour*3600 + min*60 + sec

	startH, startM, startS := windowStart.Clock()
	startSeconds := startH*3600 + startM*60 + startS

	endH, endM, endS := windowEnd.Clock()
	endSeconds := endH*3600 + endM*60 + endS

	if startSeconds <= endSeconds {
		return nowSeconds >= startSeconds && nowSeconds <= endSeconds
	}
	// Overnight window (e.g., 22:00 - 06:00)
	return nowSeconds >= startSeconds || nowSeconds <= endSeconds
}
