
.DEFAULT_GOAL := build

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(firstword $(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

setup: ## Install nlopes/slack
	go get -u github.com/nlopes/slack

fmt: ## Format all code
	go fmt *.go

build: fmt ## Build the go program
	go build .

clean: ## Remove artiacts
	rm -f main slack_emoji_grabber

run: build ## Run program
	./slack_emoji_grabber

.PHONY: help fmt clean
