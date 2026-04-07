package tee

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://localhost:8080", true)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.baseURL != "https://localhost:8080" {
		t.Errorf("expected baseURL https://localhost:8080, got %s", client.baseURL)
	}
}

func TestClientPingOffline(t *testing.T) {
	// Test ping when service is not available
	client := NewClient("https://localhost:9999", true)
	err := client.Ping()
	if err == nil {
		t.Error("expected error when pinging non-existent service")
	}
}
