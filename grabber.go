package main

import (
	"net/http"
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
