package ethutil

import (
	"context"
	"time"
	"github.com/ethereum/go-ethereum/event"
)

// WithResubscription wraps the subscribe function to call it repeatedly
// to keep a subscription alive. When a subscription is established, it is
// monitored and in the case of a failure, resubscribe is attempted by
// calling the subscribe function again.
//
// The mechanism applies backoff between resubscription attempts.
// The time between calls is adapted based on the error rate, but will never
// exceed backoffMax.
//
// The mechanism monitors the time elapsed between resubscription attempts and
// if it is shorter than the specificed alertThreshold, it writes the time
// elapsed between resubscription attempts to the thresholdViolated channel to
// alarm about potential problems with the stability of the subscription.
//
// Every error returned by the wrapped subscription function, is written to
// subscriptionFailed channel.
//
// If the calling code is interested in reading from thresholdViolated and/or
// subscriptionFailed channel, appropriate readers need to be set up *before*
// WithResubscription is called.
//
// Writes to thresholdViolated and subscriptionFailed channels are non-blocking
// and are not stopping resubscription attempts.
func WithResubscription(
	backoffMax time.Duration,
	resubscribeFn event.ResubscribeFunc,	
	alertThreshold time.Duration,
	thresholdViolated chan<- time.Duration,
	subscriptionFailed chan<- error,
) event.Subscription {
	lastAttempt := time.Time{}
	wrappedResubscribeFn := func(ctx context.Context) (event.Subscription, error) {
		now := time.Now()
		elapsed := now.Sub(lastAttempt)
		if elapsed < alertThreshold {
			select {
			case thresholdViolated <- elapsed: 
			default:
			}
		}
		
		lastAttempt = now

		sub, err := resubscribeFn(ctx)
		if err != nil {
			select {
			case subscriptionFailed <- err:
			default:
			}
		}
		return sub, err
	}

	return event.Resubscribe(backoffMax, wrappedResubscribeFn)
}
