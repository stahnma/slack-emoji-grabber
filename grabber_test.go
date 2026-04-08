package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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
		_, _ = w.Write([]byte("fake-image-data"))
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

func TestRun_DownloadsEmojis(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("image-bytes"))
	}))
	defer server.Close()

	client := &mockSlackClient{
		emojis: map[string]string{
			"partyparrot": server.URL + "/partyparrot.gif",
			"thumbsup":    server.URL + "/thumbsup.png",
		},
	}

	dir := t.TempDir()
	g := NewGrabber(client)
	g.OutputDir = dir

	if err := g.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != 2 {
		t.Errorf("expected 2 files, got %d", len(entries))
	}
}

func TestRun_SkipsAliases(t *testing.T) {
	client := &mockSlackClient{
		emojis: map[string]string{
			"myalias": "alias:partyparrot",
		},
	}

	dir := t.TempDir()
	g := NewGrabber(client)
	g.OutputDir = dir

	if err := g.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("expected 0 files (alias skipped), got %d", len(entries))
	}
}

func TestRun_SkipsExistingFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("new-data"))
	}))
	defer server.Close()

	client := &mockSlackClient{
		emojis: map[string]string{
			"existing": server.URL + "/existing.png",
		},
	}

	dir := t.TempDir()
	g := NewGrabber(client)
	g.OutputDir = dir

	// Pre-create the file
	if err := os.WriteFile(filepath.Join(dir, "existing.png"), []byte("old-data"), 0644); err != nil {
		t.Fatalf("writing test fixture: %v", err)
	}

	if err := g.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// File should NOT be overwritten
	data, _ := os.ReadFile(filepath.Join(dir, "existing.png"))
	if string(data) != "old-data" {
		t.Error("existing file was overwritten")
	}
}

func TestRun_APIError(t *testing.T) {
	client := &mockSlackClient{
		err: errors.New("slack api error"),
	}

	g := NewGrabber(client)
	g.OutputDir = t.TempDir()

	err := g.Run()
	if err == nil {
		t.Fatal("expected error from API failure")
	}
	if !strings.Contains(err.Error(), "slack api error") {
		t.Errorf("error should contain API error, got: %v", err)
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
