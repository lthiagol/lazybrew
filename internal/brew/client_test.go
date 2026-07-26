package brew

import (
	"testing"
)

func newMockClient() *Client {
	runner := NewMockRunner()
	return NewClient(runner)
}

func TestNewMockClient(t *testing.T) {
	client := newMockClient()
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if client.Formulae == nil {
		t.Error("Formulae reader should not be nil")
	}
	if client.FormulaeWrite == nil {
		t.Error("Formulae writer should not be nil")
	}
	if client.Casks == nil {
		t.Error("Casks reader should not be nil")
	}
	if client.CasksWrite == nil {
		t.Error("Casks writer should not be nil")
	}
	if client.Taps == nil {
		t.Error("Taps service should not be nil")
	}
	if client.Services == nil {
		t.Error("Services service should not be nil")
	}
	if client.Search == nil {
		t.Error("Search service should not be nil")
	}
	if client.Trust == nil {
		t.Error("Trust service should not be nil")
	}
	if client.Diagnostics == nil {
		t.Error("Diagnostics reader should not be nil")
	}
	if client.DiagnosticsWrite == nil {
		t.Error("Diagnostics writer should not be nil")
	}
	if client.Cache == nil {
		t.Error("Cache should not be nil")
	}
}

func TestNewMockClientSharesCache(t *testing.T) {
	client := newMockClient()
	if client.Cache == nil {
		t.Fatal("cache is nil")
	}

	client.Cache.Set("test-key", "test-value")
	val, ok := client.Cache.Get("test-key")
	if !ok {
		t.Fatal("expected cache to be shared across all services")
	}
	if val.(string) != "test-value" {
		t.Errorf("got %v, want test-value", val)
	}
}
