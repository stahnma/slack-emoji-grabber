package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
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

func (g *Grabber) downloadFile(fpath, url string) error {
	resp, err := g.HTTPClient.Get(url)
	if err != nil {
		return fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}

	out, err := os.Create(fpath)
	if err != nil {
		return fmt.Errorf("creating %s: %w", fpath, err)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		os.Remove(fpath)
		return fmt.Errorf("writing %s: %w", fpath, err)
	}

	return out.Close()
}
