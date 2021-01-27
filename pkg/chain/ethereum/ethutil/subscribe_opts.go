package ethutil

import "time"

const (
	// DefaultSubscribeOptsTick is the default duration with which
	// past events are pulled from the chain by the subscription monitoring
	// mechanism if no other value is provided in SubscribeOpts when creating
	// the subscription.
	DefaultSubscribeOptsTick = 15 * time.Minute

	// DefaultSubscribeOptsBlocksBack is the default number of past blocks
	// pulled from the chain by the subscription monitoring mechanism if no
	// other value is provided in SubscribeOpts when creating the subscription.
	DefaultSubscribeOptsBlocksBack = 100
)

// SubscribeOpts specifies optional configuration options that can be passed
// when creating Ethereum event subscription.
type SubscribeOpts struct {

	// Tick is the duration with which subscription monitoring mechanism
	// pulls events from the chain. This mechanism is an additional process
	// next to a regular watchLogs subscription making sure no events are lost
	// even in case the regular subscription missed them because of, for
	// example, connectivity problems.
	Tick time.Duration

	// BlocksBack is the number of past blocks subscription monitoring mechanism
	// takes into consideration when pulling past events from the chain.
	// This event pull mechanism is an additional process next to a regular
	// watchLogs subscription making sure no events are lost even in case the
	// regular subscription missed them because of, for example, connectivity
	// problems.
	BlocksBack uint64
}
