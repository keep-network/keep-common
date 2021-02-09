package ethlike

import (
	"context"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
	"time"
)

// RateLimiter is a helper tool which allows controlling the number and
// concurrency of requests made against a generic target.
type RateLimiter struct {
	limiter              *rate.Limiter
	semaphore            *semaphore.Weighted
	acquirePermitTimeout time.Duration
}

// RateLimiterConfig represents the configuration of the rate limiter.
type RateLimiterConfig struct {
	// RequestsPerSecondLimit sets the maximum average number of requests
	// per second. It's important to note that in short periods of time
	// the actual average may exceed this limit slightly.
	RequestsPerSecondLimit int

	// ConcurrencyLimit sets the maximum number of concurrent requests which
	// can be executed against the target at the same time.
	ConcurrencyLimit int

	// AcquirePermitTimeout determines how long a request can wait trying
	// to acquire a permit from the rate limiter.
	AcquirePermitTimeout time.Duration
}

// NewRateLimiter creates a new rate limiter instance basing on given config.
func NewRateLimiter(
	config *RateLimiterConfig,
) *RateLimiter {
	rateLimiter := &RateLimiter{}

	if config.RequestsPerSecondLimit > 0 {
		rateLimiter.limiter = rate.NewLimiter(
			rate.Limit(config.RequestsPerSecondLimit),
			1,
		)
	}

	if config.ConcurrencyLimit > 0 {
		rateLimiter.semaphore = semaphore.NewWeighted(
			int64(config.ConcurrencyLimit),
		)
	}

	if config.AcquirePermitTimeout > 0 {
		rateLimiter.acquirePermitTimeout = config.AcquirePermitTimeout
	} else {
		rateLimiter.acquirePermitTimeout = 5 * time.Minute
	}

	return rateLimiter
}

func (rl *RateLimiter) AcquirePermit() error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		rl.acquirePermitTimeout,
	)
	defer cancel()

	if rl.limiter != nil {
		err := rl.limiter.Wait(ctx)
		if err != nil {
			return err
		}
	}

	if rl.semaphore != nil {
		err := rl.semaphore.Acquire(ctx, 1)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rl *RateLimiter) ReleasePermit() {
	if rl.semaphore != nil {
		rl.semaphore.Release(1)
	}
}
