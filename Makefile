export CGO_ENABLED := 0
BINARY := slack-emoji-grabber
MODULE := github.com/stahnma/slack-emoji-grabber
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
CMD := ./cmd/slack-emoji-grabber

.PHONY: build clean test lint vet fmt run install help

build: ## Build the binary
	go build $(LDFLAGS) -o $(BINARY) $(CMD)

install: ## Install to GOPATH/bin
	go install $(LDFLAGS) $(CMD)

test: ## Run tests
	go test ./... -v

test-short: ## Run tests (short mode)
	go test ./... -short

vet: ## Run go vet
	go vet ./...

fmt: ## Run gofmt
	gofmt -s -w .

lint: vet fmt ## Run all linters

tidy: ## Run go mod tidy
	go mod tidy

clean: ## Remove build artifacts
	rm -f $(BINARY)
	go clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
