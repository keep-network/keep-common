package ethutil

import "time"

const (
	DefaultSubscribeOptsTickDuration = 15 * time.Minute
	DefaultSubscribeOptsBlocksBack   = 100
)

type SubscribeOpts struct {
	TickDuration time.Duration
	BlocksBack   uint64
}
