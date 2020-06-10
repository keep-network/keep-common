package metrics

import (
	"context"
	"time"
)

// Source defines a source of metric data.
type Source func() float64

// Sink defines a destination of collected metric data.
type Sink interface {
	Set(value float64)
}

// Observe triggers a cyclic metric observation goroutine.
func Observe(
	ctx context.Context,
	source Source,
	sink Sink,
	tick time.Duration,
) {
	go func() {
		ticker := time.NewTicker(tick)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				sink.Set(source())
			case <-ctx.Done():
				return
			}
		}
	}()
}
