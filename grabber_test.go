package main

import (
	"testing"
)

// mockSlackClient implements SlackClient for testing
type mockSlackClient struct {
	emojis map[string]string
	err    error
}

func (m *mockSlackClient) GetEmoji() (map[string]string, error) {
	return m.emojis, m.err
}

func TestNewGrabber(t *testing.T) {
	client := &mockSlackClient{}
	g := NewGrabber(client)
	if g.OutputDir != "emojis" {
		t.Errorf("expected default output dir 'emojis', got %q", g.OutputDir)
	}
	if g.Client == nil {
		t.Error("expected non-nil client")
	}
	if g.HTTPClient == nil {
		t.Error("expected non-nil HTTP client")
	}
}
