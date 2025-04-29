package files

import "testing"

func TestMetricConfig_Get(t *testing.T) {

	var expectedKeys = []string{
		"$from",
		"$metric",
		"$to",
	}

	mc := NewMetricConfig()
	result, err := mc.Get("binary")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.String("$metric") != "binary" {
		t.Fatalf("Expected metric to be 'binary', got %s", result.String("$metric"))
	}
	if len(result) != len(expectedKeys) {
		t.Fatalf("Expected %d keys, got %d", len(expectedKeys), len(result))
	}

	expectedSubject := "V1.gauge.nodeQ2EdemoQ2Eat.hostQ2EdemoQ2Eat.aQ20serviceQ20name"
	expectedKeys = []string{
		"$from",
		"$metric",
		"$nats_subject",
		"$to",
	}

	mc = NewMetricConfig().WithMacros(map[string]any{
		"k": "gauge",
		"n": "node.demo.at",
		"h": "host.demo.at",
		"s": "a service name",
	})

	result, err = mc.Get("gauge")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result) != len(expectedKeys) {
		t.Fatalf("Expected %d keys, got %d", len(expectedKeys), len(result))
	}

	if result.String("$nats_subject") != expectedSubject {
		t.Fatalf("Expected %q, got %q", expectedSubject, result.String("$metric"))
	}
}
