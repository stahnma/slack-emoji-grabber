package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/slack-go/slack"
)

var version = "dev"

func main() {
	outputDir := flag.String("output", "emojis", "directory to save emojis")
	showVersion := flag.Bool("version", false, "print version and exit")
	verbose := flag.Bool("v", false, "enable verbose/debug logging")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	if *verbose {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}

	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		log.Fatal("SLACK_TOKEN environment variable is required")
	}

	g := NewGrabber(slack.New(token))
	g.OutputDir = *outputDir

	if err := g.Run(); err != nil {
		log.Fatal(err)
	}
}
