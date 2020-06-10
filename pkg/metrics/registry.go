// Package `metrics` provides some tools useful for gathering and
// exposing system metrics for external monitoring tools.
//
// Currently, this package is intended to use with Prometheus but
// can be easily extended if needed. Also, not all Prometheus metric
// types are implemented. The main motivation of creating a custom
// package was a need to avoid usage of external unaudited dependencies.
//
// Following specifications was used as reference:
// - https://prometheus.io/docs/instrumenting/writing_clientlibs/
// - https://prometheus.io/docs/instrumenting/exposition_formats/
package metrics

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/ipfs/go-log"
)

var logger = log.Logger("keep-metrics")

type metric interface {
	expose() string
}

// Registry performs all management of metrics. Specifically, it allows
// to registering new metrics and exposing them through the metrics server.
type Registry struct {
	application string
	identifier  string

	metrics      map[string]metric
	metricsMutex sync.RWMutex
}

// NewRegistry creates a new metrics registry.
func NewRegistry(application, identifier string) *Registry {
	return &Registry{
		application: application,
		identifier:  identifier,
		metrics:     make(map[string]metric),
	}
}

// EnableServer enables the metrics server on the given port. Data will
// be exposed on `/metrics` path.
func (r *Registry) EnableServer(port int) {
	server := &http.Server{Addr: ":" + strconv.Itoa(port)}

	http.HandleFunc("/metrics", func(response http.ResponseWriter, _ *http.Request) {
		if _, err := io.WriteString(response, r.exposeMetrics()); err != nil {
			logger.Errorf("could not write response: [%v]", err)
		}
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Errorf("metrics server error: [%v]", err)
		}
	}()
}

// Exposes all registered metrics in their text format.
func (r *Registry) exposeMetrics() string {
	r.metricsMutex.RLock()
	defer r.metricsMutex.RUnlock()

	metrics := make([]string, 0)
	for _, metric := range r.metrics {
		metrics = append(metrics, metric.expose())
	}

	return strings.Join(metrics, "\n\n")
}

// NewGauge creates and registers a new gauge metric which will be exposed
// through the metrics server. In case a metric already exists, an error
// will be returned.
func (r *Registry) NewGauge(name string) (*Gauge, error) {
	r.metricsMutex.Lock()
	defer r.metricsMutex.Unlock()

	if _, exists := r.metrics[name]; exists {
		return nil, fmt.Errorf("gauge [%v] already exists", name)
	}

	gauge := &Gauge{
		name: name,
		labels: map[string]string{
			"application": r.application,
			"identifier":  r.identifier,
		},
	}

	r.metrics[name] = gauge
	return gauge, nil
}
