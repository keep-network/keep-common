package ethutil

import (
	"context"
	"time"
	"github.com/ethereum/go-ethereum/event"
)

// WithResubscription wraps the resubscription function to call it repeatedly
// to keep a subscription established. When subscription is established, it is
// monitored and in the case of a failure, resubscription is attempted by
// calling the function again.
//
// The mechanism applies backoff between calls to resubscribe function.
// The time between calls is adapted based on the error rate, but will never
// exceed backoffMax.
//
// The mechanism monitors the time elapsed between resubscription attempts and
// if it is shorter than the specificed alertThreshold, it calls the alertFn
// to alarm about potential problems with the stability of the subscription.
func WithResubscription(
	backoffMax time.Duration,
	resubscribeFn event.ResubscribeFunc,
	alertThreshold time.Duration,
	alertFn func(),
) event.Subscription {
	lastAttempt := time.Time{}
	wrappedResubscribeFn := func(ctx context.Context) (event.Subscription, error) {
		now := time.Now()
		elapsed := now.Sub(lastAttempt)
		if elapsed < alertThreshold {
			alertFn()
		}
		lastAttempt = now
		return resubscribeFn(ctx)
	}

	return event.Resubscribe(backoffMax, wrappedResubscribeFn)
}
