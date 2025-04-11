package files

import "testing"

func TestBaseConfig_GetString(t *testing.T) {
	bc := newBaseConfig()
	apiKey, err := bc.getString("ApiKey")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if apiKey == "" {
		t.Fatalf("API key is empty")
	}
	mc := NewMetricConfig()
	apiKey1, err := mc.GetFromBase("ApiKey")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if apiKey1 == "" {
		t.Fatalf("API key is empty")
	}
	if apiKey != apiKey1 {
		t.Fatalf("API keys do not match: %s != %s", apiKey, apiKey1)
	}
}
