package downloader

import (
	"context"
	"math/rand"
	"time"
)

const (
	defaultMaxRetries = 5
	defaultBaseDelay  = 500 * time.Millisecond
	defaultMaxDelay   = 10 * time.Second
)

func calculateBackOff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	multiplier := 1 << attempt // 2^ attempt: 1, 2, 4, 8, ...
	backOff := min(baseDelay*time.Duration(multiplier), maxDelay)

	// Add full jitter (0 to 50% of the backoff) to prevent thundering herds
	jitter := time.Duration(rand.Int63n(int64(backOff / 2)))
	return backOff/2 + jitter
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}
