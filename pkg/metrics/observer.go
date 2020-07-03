package metrics

import (
	"context"
	"time"
)

// ObserverInput defines a source of metric data.
type ObserverInput func() float64

// ObserverOutput defines a destination of collected metric data.
type ObserverOutput interface {
	Set(value float64)
}

// Observer represent a definition of a cyclic metric observation process.
type Observer struct {
	input  ObserverInput
	output ObserverOutput
}

// Observe triggers a cyclic metric observation process.
func (o *Observer) Observe(
	ctx context.Context,
	tick time.Duration,
) {
	go func() {
		o.output.Set(o.input()) // execute the first check immediately

		ticker := time.NewTicker(tick)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				o.output.Set(o.input())
			case <-ctx.Done():
				return
			}
		}
	}()
}
