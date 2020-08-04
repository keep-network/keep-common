// Package `diagnostics` provides some tools useful for gathering and
// exposing arbitrary diagnositcs information for external monitoring tools.
//
// Possible usage: integration nodes list into dashboard
package diagnostics

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/ipfs/go-log"
)

var logger = log.Logger("keep-diagnostics")

// Registry performs all management of diagnostic. Specifically, it allows
// to registering new diagnostics sources and exposing them through the diagnostics server.
type DiagnosticsRegistry struct {
	diagnosticsSources map[string]func() string
	diagnosticsMutex   sync.RWMutex
}

// NewRegistry creates a new metrics registry.
func NewRegistry() *DiagnosticsRegistry {
	return &DiagnosticsRegistry{
		diagnosticsSources: make(map[string]func() string),
	}
}

// EnableServer enables the diagnostics server on the given port. Data will
// be exposed on `/diagnostics` path in JSON format.
func (r *DiagnosticsRegistry) EnableServer(port int) {
	server := &http.Server{Addr: ":" + strconv.Itoa(port)}

	http.HandleFunc("/diagnostics", func(response http.ResponseWriter, _ *http.Request) {
		if _, err := io.WriteString(response, r.exposeDiagnostics()); err != nil {
			logger.Errorf("could not write response: [%v]", err)
		}
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Errorf("diagnostics server error: [%v]", err)
		}
	}()
}

func (r *DiagnosticsRegistry) RegisterSource(name string, source func() string) {
	r.diagnosticsMutex.Lock()
	defer r.diagnosticsMutex.Unlock()

	r.diagnosticsSources[name] = source
}

// Exposes all registered diagnostics sources in a single JSON document.
func (r *DiagnosticsRegistry) exposeDiagnostics() string {
	r.diagnosticsMutex.RLock()
	defer r.diagnosticsMutex.RUnlock()

	diagnostics := make(map[string]interface{})
	for sourceName, sourceGetter := range r.diagnosticsSources {
		var jsonString = sourceGetter()
		var jsonObject interface{}
		err := json.Unmarshal([]byte(jsonString), &jsonObject)
		if err == nil {
			diagnostics[sourceName] = jsonObject
		}
	}

	bytes, err := json.Marshal(diagnostics)
	if err != nil {
		return ""
	}

	return string(bytes)
}
