package ethutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/event"
)

func TestResubscribeAboveThreshold(t *testing.T) {
	backoffMax := 100 * time.Millisecond
	alertThreshold := 100 * time.Millisecond

	alertFnCalls := 0
	alertFn := func() {
		alertFnCalls++
	}

	subscriptionFailures := 3
	resubscribeFnCalls := 0
	resubscribeFn := func(ctx context.Context) (event.Subscription, error) {
		resubscribeFnCalls++
		time.Sleep(150 * time.Millisecond) // 150 > 100, above alert threshold
		if resubscribeFnCalls <= subscriptionFailures {
			return nil, fmt.Errorf("this is the way")
		}
		delegate := event.NewSubscription(func(unsubscribed <-chan struct{}) error {
			return nil
		})
		return delegate, nil
	}

	subscription := WithResubscription(
		backoffMax,
		resubscribeFn,
		alertThreshold,
		alertFn,
	)
	<-subscription.Err()

	// No calls to alertFn expected. Alert threshold is set to 100ms and no
	// there were no resubscription attempts in a time shorter than 150ms one
	// after another.
	if alertFnCalls != 0 {
		t.Fatalf("alert triggered [%v] times, expected none", alertFnCalls)
	}

	// Subscription failed `subscriptionFailures` times and resubscription
	// function should be called `subscriptionFailures + 1` times - one time
	// for each failure and one time at the end - that subscription was
	// successful and had not to be retried.
	expectedResubscriptionCalls := subscriptionFailures + 1
	if resubscribeFnCalls != expectedResubscriptionCalls {
		t.Fatalf(
			"resubscription called [%v] times, expected [%v]",
			resubscribeFnCalls,
			expectedResubscriptionCalls,
		)
	}
}

func TestResubscribeBelowThreshold(t *testing.T) {
	backoffMax := 50 * time.Millisecond
	alertThreshold := 100 * time.Millisecond

	alertFnCalls := 0
	alertFn := func() {
		alertFnCalls++
	}

	subscriptionFailures := 5
	resubscribeFnCalls := 0
	resubscribeFn := func(ctx context.Context) (event.Subscription, error) {
		resubscribeFnCalls++
		time.Sleep(50 * time.Millisecond) // 50 < 100, below alert threshold
		if resubscribeFnCalls <= subscriptionFailures {
			return nil, fmt.Errorf("i have spoken")
		}
		delegate := event.NewSubscription(func(unsubscribed <-chan struct{}) error {
			return nil
		})
		return delegate, nil
	}

	subscription := WithResubscription(
		backoffMax,
		resubscribeFn,
		alertThreshold,
		alertFn,
	)
	<-subscription.Err()

	fmt.Printf("resubscribe count = [%v]\n", resubscribeFnCalls)
	// Alert function should be called for each subscription failure if the time
	// between the previous resubscription was shorter than the threshold.
	// In this test, alert threshold is set to 100ms and delays between failures
	// are just 50ms.
	expectedAlertFnCalls := subscriptionFailures
	if alertFnCalls != expectedAlertFnCalls {
		t.Fatalf(
			"alert triggered [%v] times, expected [%v]",
			alertFnCalls,
			expectedAlertFnCalls,
		)
	}

	// Subscription failed `subscriptionFailures` times and resubscription
	// function should be called `subscriptionFailures + 1` times - one time
	// for each failure and one time at the end - that subscription was
	// successful and had not to be retried.
	expectedResubscriptionCalls := subscriptionFailures + 1
	if resubscribeFnCalls != expectedResubscriptionCalls {
		t.Fatalf(
			"resubscription called [%v] times, expected [%v]",
			resubscribeFnCalls,
			expectedResubscriptionCalls,
		)
	}
}
