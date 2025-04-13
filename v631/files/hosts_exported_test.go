package files

import "testing"

func TestHostsExported_GetHostProperties(t *testing.T) {
	_, err := NewHostsExported().GetHostProperties("host-does-not-exist")
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
}
