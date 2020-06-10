package metrics

import (
	"testing"
)

func TestRegistryNewGauge(t *testing.T) {
	registry := NewRegistry("test-app", "test-id")

	gauge, err := registry.NewGauge("test-gauge")
	if err != nil {
		t.Fatal(err)
	}

	if _, err = registry.NewGauge("test-gauge"); err == nil {
		t.Fatalf("should fail when creating gauge with the same name")
	}

	if _, exists := registry.metrics[gauge.name]; !exists {
		t.Fatalf("metric with name [%v] should exist in the registry", gauge.name)
	}
}
