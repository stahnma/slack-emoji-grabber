package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestDownloadFile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fake-image-data"))
	}))
	defer server.Close()

	g := NewGrabber(&mockSlackClient{})
	dir := t.TempDir()
	dest := filepath.Join(dir, "test.png")

	err := g.downloadFile(dest, server.URL+"/emoji.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != "fake-image-data" {
		t.Errorf("expected 'fake-image-data', got %q", string(data))
	}
}

func TestDownloadFile_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	g := NewGrabber(&mockSlackClient{})
	dir := t.TempDir()
	dest := filepath.Join(dir, "test.png")

	err := g.downloadFile(dest, server.URL+"/missing.png")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}

	// Partial file should not exist
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Error("expected partial file to be removed after HTTP error")
	}
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
