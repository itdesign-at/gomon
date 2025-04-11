package files

import "testing"

func TestMetricConfig_Get(t *testing.T) {
	mc := NewMetricConfig()
	result, err := mc.Get("binary")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.String("$metric") != "binary" {
		t.Fatalf("Expected metric to be 'binary', got %s", result.String("$metric"))
	}
}
