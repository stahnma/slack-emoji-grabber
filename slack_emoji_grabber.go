package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/slack-go/slack"
)

var version = "dev"

func downloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	slacktoken := os.Getenv("SLACK_TOKEN")
	if slacktoken == "" {
		log.Fatal("SLACK_TOKEN environment variable is required")
	}
	api := slack.New(slacktoken)
	emojiset, err := api.GetEmoji()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	// make the dir
	os.Mkdir("./emojis", 0755)
	// loop over each emoji
	for name, uri := range emojiset {
		filepath := "./emojis/" + name
		lastdotindex := strings.LastIndexAny(uri, ".")
		if lastdotindex != -1 {
			suffix := uri[lastdotindex:]
			filepath = filepath + suffix
		}
		if strings.Contains(uri, "alias:") {
			fmt.Println("Skipping " + name + " because it is an alias.")
			continue
		}
		// check to see if we have the download already
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			// if no, download
			fmt.Println(filepath)
			downloadFile(filepath, uri)
		} else {
			fmt.Println("Already found " + filepath)
		}
	}
}
