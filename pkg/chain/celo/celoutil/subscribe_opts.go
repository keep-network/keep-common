package celoutil

import "time"

const (
	// DefaultSubscribeOptsTick is the default duration with which
	// past events are pulled from the chain by the subscription monitoring
	// mechanism if no other value is provided in SubscribeOpts when creating
	// the subscription.
	DefaultSubscribeOptsTick = 15 * time.Minute

	// DefaultSubscribeOptsPastBlocks is the default number of past blocks
	// pulled from the chain by the subscription monitoring mechanism if no
	// other value is provided in SubscribeOpts when creating the subscription.
	DefaultSubscribeOptsPastBlocks = 100

	// SubscriptionBackoffMax is the maximum backoff time between event
	// resubscription attempts.
	SubscriptionBackoffMax = 2 * time.Minute

	// SubscriptionAlertThreshold is time threshold below which event
	// resubscription emits an error to the logs.
	// WS connection can be dropped at any moment and event resubscription will
	// follow. However, if WS connection for event subscription is getting
	// dropped too often, it may indicate something is wrong with Celo
	// client. This constant defines the minimum lifetime of an event
	// subscription required before the subscription failure happens and
	// resubscription follows so that the resubscription does not emit an error
	// to the logs alerting about potential problems with Celo client
	// connection.
	SubscriptionAlertThreshold = 15 * time.Minute
)

// SubscribeOpts specifies optional configuration options that can be passed
// when creating Celo event subscription.
type SubscribeOpts struct {

	// Tick is the duration with which subscription monitoring mechanism
	// pulls events from the chain. This mechanism is an additional process
	// next to a regular watchLogs subscription making sure no events are lost
	// even in case the regular subscription missed them because of, for
	// example, connectivity problems.
	Tick time.Duration

	// PastBlocks is the number of past blocks subscription monitoring mechanism
	// takes into consideration when pulling past events from the chain.
	// This event pull mechanism is an additional process next to a regular
	// watchLogs subscription making sure no events are lost even in case the
	// regular subscription missed them because of, for example, connectivity
	// problems.
	PastBlocks uint64
}
