package clientinfo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func TestClientInfoServerMetrics(t *testing.T) {
	registry := NewRegistry()
	expectedMetricsFamily := map[string]*dto.MetricFamily{}

	// Register Gauge Metric
	metricGaugeName := "test_gauge"
	metricGaugeValue := float64(12)
	metricGaugeTimestamp := int64(51241)
	metricGaugeLabelName := "gaugeLabelName"
	metricGaugeLabelValue := "gauge-label-value"

	gauge, err := registry.NewMetricGauge(metricGaugeName, NewLabel(metricGaugeLabelName, metricGaugeLabelValue))
	if err != nil {
		t.Fatal(err)
	}
	gauge.value = metricGaugeValue
	gauge.timestamp = metricGaugeTimestamp

	expectedMetricsFamily[metricGaugeName] = &dto.MetricFamily{
		Name: &metricGaugeName,
		Type: dto.MetricType_GAUGE.Enum(),
		Metric: []*dto.Metric{
			{
				Label: []*dto.LabelPair{
					{
						Name:  &metricGaugeLabelName,
						Value: &metricGaugeLabelValue,
					},
				},
				Gauge: &dto.Gauge{
					Value: &metricGaugeValue,
				},
				TimestampMs: &metricGaugeTimestamp,
			},
		},
	}

	// Register Info Metric
	metricInfoName := "test_info"
	metricInfoValue := float64(1) // Default value resolved by metric parser for Info metrics.
	metricInfoLabelName := "infoLabelName"
	metricInfoLabelValue := "info-label-value"

	if _, err := registry.NewMetricInfo(
		metricInfoName,
		[]Label{
			NewLabel(metricInfoLabelName, metricInfoLabelValue),
		},
	); err != nil {
		t.Fatal(err)
	}

	expectedMetricsFamily[metricInfoName] = &dto.MetricFamily{
		Name: &metricInfoName,
		Type: dto.MetricType_UNTYPED.Enum(),
		Metric: []*dto.Metric{
			{
				Label: []*dto.LabelPair{
					{
						Name:  &metricInfoLabelName,
						Value: &metricInfoLabelValue,
					},
				},
				Untyped: &dto.Untyped{Value: &metricInfoValue},
			},
		},
	}

	// Execute Test
	port := 9799
	registry.EnableServer(port)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", port))
	if err != nil {
		t.Fatalf("failed to get metrics: %v", err)
	}

	var parser expfmt.TextParser
	metrics, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		t.Fatalf("failed to parse metrics: %v", err)
	}

	if !reflect.DeepEqual(expectedMetricsFamily, metrics) {
		t.Errorf(
			"incorrect metrics family\n"+
				"expected: %+v\n"+
				"actual:   %+v",
			expectedMetricsFamily,
			metrics,
		)
	}
}

func TestClientInfoServerDiagnostics(t *testing.T) {
	registry := NewRegistry()

	// Register Diagnostics Sources
	nodeChainAddress := "0x1234567890"
	registry.RegisterDiagnosticSource("chain_address", func() string {
		result, err := json.Marshal(nodeChainAddress)
		if err != nil {
			t.Fatal(err)
		}
		return string(result)
	})

	peers := []testDiagnosticsPeerInfo{
		{
			ChainAddress:          "0xABCdef",
			NetworkMultiAddresses: []string{"/dns4/address1/3919/ipfs/AaBbCc"},
		},
		{
			ChainAddress:          "0x98765A",
			NetworkMultiAddresses: []string{"/ip4/127.0.0.1/3919/ipfs/963"},
		},
	}
	registry.RegisterDiagnosticSource("connected_peers", func() string {
		result, err := json.Marshal(peers)
		if err != nil {
			t.Fatal(err)
		}
		return string(result)
	})

	expectedDiagnostics := &testDiagnosticsInfo{nodeChainAddress, peers}

	// Execute Test
	port := 9899
	registry.EnableServer(port)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/diagnostics", port))
	if err != nil {
		t.Fatalf("failed to get diagnostics: %v", err)
	}

	actualDiagnostics := &testDiagnosticsInfo{}
	if err = json.NewDecoder(resp.Body).Decode(actualDiagnostics); err != nil {
		t.Fatalf("failed to decode diagnostics: %v", err)
	}

	if !reflect.DeepEqual(expectedDiagnostics, actualDiagnostics) {
		t.Errorf(
			"incorrect diagnostics\n"+
				"expected: %+v\n"+
				"actual:   %+v",
			expectedDiagnostics,
			actualDiagnostics,
		)
	}
}

type testDiagnosticsInfo struct {
	ChainAddress   string                    `json:"chain_address"`
	ConnectedPeers []testDiagnosticsPeerInfo `json:"connected_peers"`
}

type testDiagnosticsPeerInfo struct {
	ChainAddress          string   `json:"chain_address"`
	NetworkMultiAddresses []string `json:"multiaddrs"`
}
