package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"

	grabber "github.com/stahnma/slack-emoji-grabber"

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	g := grabber.NewGrabber(slack.New(token))
	g.OutputDir = *outputDir

	if err := g.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
