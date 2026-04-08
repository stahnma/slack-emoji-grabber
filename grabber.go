package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// SlackClient abstracts the Slack API for testability.
type SlackClient interface {
	GetEmoji() (map[string]string, error)
}

// Grabber holds configuration for downloading emojis.
type Grabber struct {
	Client     SlackClient
	HTTPClient *http.Client
	OutputDir  string
}

// NewGrabber creates a Grabber with sensible defaults.
func NewGrabber(client SlackClient) *Grabber {
	return &Grabber{
		Client:     client,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		OutputDir:  "emojis",
	}
}

// Run fetches the emoji list and downloads each one.
func (g *Grabber) Run() error {
	emojis, err := g.Client.GetEmoji()
	if err != nil {
		return fmt.Errorf("fetching emoji list: %w", err)
	}

	if err := os.MkdirAll(g.OutputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	for name, uri := range emojis {
		if strings.HasPrefix(uri, "alias:") {
			slog.Debug("skipping alias", "name", name)
			continue
		}

		ext := path.Ext(uri)
		fpath := filepath.Join(g.OutputDir, name+ext)

		if _, err := os.Stat(fpath); err == nil {
			slog.Debug("skipping existing", "path", fpath)
			continue
		}

		slog.Info("downloading", "name", name, "path", fpath)
		if err := g.downloadFile(fpath, uri); err != nil {
			slog.Error("download failed", "name", name, "err", err)
			continue
		}
	}

	return nil
}

func (g *Grabber) downloadFile(fpath, url string) error {
	resp, err := g.HTTPClient.Get(url)
	if err != nil {
		return fmt.Errorf("fetching %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}

	out, err := os.Create(fpath)
	if err != nil {
		return fmt.Errorf("creating %s: %w", fpath, err)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		_ = out.Close()
		_ = os.Remove(fpath)
		return fmt.Errorf("writing %s: %w", fpath, err)
	}

	return out.Close()
}
