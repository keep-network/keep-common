// Package metrics provides some tools useful for gathering and
// exposing system metrics for external monitoring tools.
//
// Currently, this package is intended to use with Prometheus but
// can be easily extended if needed. Also, not all Prometheus metric
// types are implemented.
//
// Following specifications were used as reference:
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

// Label represents an arbitrary information attached to the metrics.
type Label struct {
	name  string
	value string
}

// NewLabel creates a new label using the given name and value.
func NewLabel(name, value string) Label {
	return Label{name, value}
}

// Registry performs all management of metrics. Specifically, it allows
// to registering new metrics and exposing them through the metrics server.
type Registry struct {
	metrics      map[string]metric
	metricsMutex sync.RWMutex
}

// NewRegistry creates a new metrics registry.
func NewRegistry() *Registry {
	return &Registry{
		metrics: make(map[string]metric),
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
func (r *Registry) NewGauge(
	name string,
	labels ...Label,
) (*Gauge, error) {
	r.metricsMutex.Lock()
	defer r.metricsMutex.Unlock()

	if _, exists := r.metrics[name]; exists {
		return nil, fmt.Errorf("metric [%v] already exists", name)
	}

	gauge := &Gauge{
		name:   name,
		labels: processLabels(labels),
	}

	r.metrics[name] = gauge
	return gauge, nil
}

// NewGaugeObserver creates and registers a gauge just like `NewGauge` method
// and wrap it with a ready to use observer of the provided input. This allows
// to easily create self-refreshing metrics.
func (r *Registry) NewGaugeObserver(
	name string,
	input ObserverInput,
	labels ...Label,
) (*Observer, error) {
	gauge, err := r.NewGauge(name, labels...)
	if err != nil {
		return nil, err
	}

	return &Observer{
		input:  input,
		output: gauge,
	}, nil
}

// NewInfo creates and registers a new info metric which will be exposed
// through the metrics server. In case a metric already exists, an error
// will be returned.
func (r *Registry) NewInfo(
	name string,
	labels []Label,
) (*Info, error) {
	r.metricsMutex.Lock()
	defer r.metricsMutex.Unlock()

	if _, exists := r.metrics[name]; exists {
		return nil, fmt.Errorf("metric [%v] already exists", name)
	}

	if len(labels) == 0 {
		return nil, fmt.Errorf("at least one label should be set")
	}

	info := &Info{
		name:   name,
		labels: processLabels(labels),
	}

	r.metrics[name] = info
	return info, nil
}

func processLabels(
	labels []Label,
) map[string]string {
	result := make(map[string]string)

	for _, label := range labels {
		if label.name == "" || label.value == "" {
			continue
		}

		result[label.name] = label.value
	}

	return result
}
