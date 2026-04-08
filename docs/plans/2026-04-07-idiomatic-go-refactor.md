# Idiomatic Go Refactor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refactor slack-emoji-grabber into idiomatic, testable, robust Go with proper error handling, CLI flags, structured logging, and full test coverage.

**Architecture:** Extract core logic from `main()` into a testable `run()` function. Introduce a `Grabber` struct that holds configuration (output dir, HTTP client, Slack client interface). Use `flag` for CLI options. Use `log/slog` for structured logging. Use `net/http` test server for download tests and an interface for the Slack API to enable unit testing without live credentials.

**Tech Stack:** Go 1.25+, `log/slog`, `flag`, `path/filepath`, `net/http/httptest`, `testing`, `github.com/slack-go/slack`

---

## Task 1: Fix module path mismatch

The `go.mod` module path uses an underscore (`slack_emoji_grabber`) but the repo and Makefile use a hyphen (`slack-emoji-grabber`). This causes build/install mismatches.

**Files:**
- Modify: `go.mod:1`

**Step 1: Fix the module path**

Change line 1 of `go.mod` from:
```
module github.com/stahnma/slack_emoji_grabber
```
to:
```
module github.com/stahnma/slack-emoji-grabber
```

**Step 2: Verify it builds**

Run: `go build -o slack-emoji-grabber .`
Expected: Clean build, no errors.

**Step 3: Commit**

```bash
git add go.mod
git commit -m "fix: correct module path to use hyphen, matching repo name"
```

---

## Task 2: Add version variable and `-version` flag

The Makefile injects `-X main.version=$(VERSION)` via ldflags, but no `version` variable exists in the source. Add the variable, add `flag` parsing, and add a `-version` flag.

**Files:**
- Modify: `slack_emoji_grabber.go`

**Step 1: Add version var, flag parsing, and `-version` flag**

Add at the top of the file (after imports):

```go
var version = "dev"
```

Replace the current `main()` function opening (lines 32-37) with:

```go
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
```

Add `"flag"` and `"log"` to the imports. Remove `"fmt"` only if no longer used (it will still be used for version printing).

**Step 2: Verify build with ldflags**

Run: `go build -ldflags "-X main.version=v0.0.1-test" -o slack-emoji-grabber . && ./slack-emoji-grabber -version`
Expected output: `v0.0.1-test`

**Step 3: Commit**

```bash
git add slack_emoji_grabber.go
git commit -m "feat: add version variable and -version flag for ldflags injection"
```

---

## Task 3: Define SlackClient interface and Grabber struct

Create the core abstraction layer. The `SlackClient` interface wraps the Slack API call so we can mock it in tests. The `Grabber` struct holds all configuration.

**Files:**
- Create: `grabber.go`

**Step 1: Write the failing test**

Create `grabber_test.go`:

```go
package main

import (
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
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestNewGrabber -v`
Expected: FAIL — `NewGrabber` and types not defined.

**Step 3: Write the implementation**

Create `grabber.go`:

```go
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
```

**Step 4: Run test to verify it passes**

Run: `go test -run TestNewGrabber -v`
Expected: PASS

**Step 5: Commit**

```bash
git add grabber.go grabber_test.go
git commit -m "feat: add SlackClient interface, Grabber struct, and NewGrabber constructor"
```

---

## Task 4: Implement `downloadFile` as a Grabber method with proper error handling

Move `downloadFile` from a standalone function to a method on `Grabber`. Add HTTP status checking, error wrapping, and partial-file cleanup.

**Files:**
- Modify: `grabber.go`
- Modify: `grabber_test.go`

**Step 1: Write the failing tests**

Append to `grabber_test.go`:

```go
import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

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

	// Partial file should be cleaned up
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Error("expected partial file to be removed after HTTP error")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test -run TestDownloadFile -v`
Expected: FAIL — `g.downloadFile` not defined as method.

**Step 3: Write the implementation**

Add to `grabber.go`:

```go
import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

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
```

**Step 4: Run tests to verify they pass**

Run: `go test -run TestDownloadFile -v`
Expected: PASS

**Step 5: Commit**

```bash
git add grabber.go grabber_test.go
git commit -m "feat: move downloadFile to Grabber method with HTTP status checks and error wrapping"
```

---

## Task 5: Implement `Grabber.Run()` — the core logic

Extract the emoji-fetching loop from `main()` into a testable `Run()` method on `Grabber`. This handles: creating the output dir, iterating emojis, skipping aliases, skipping existing files, and downloading.

**Files:**
- Modify: `grabber.go`
- Modify: `grabber_test.go`

**Step 1: Write the failing tests**

Append to `grabber_test.go`:

```go
import (
	"errors"
	// ... existing imports plus:
	"strings"
)

func TestRun_DownloadsEmojis(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("image-bytes"))
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

	// Both files should exist
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
		w.Write([]byte("new-data"))
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
	os.WriteFile(filepath.Join(dir, "existing.png"), []byte("old-data"), 0644)

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
```

**Step 2: Run tests to verify they fail**

Run: `go test -run TestRun -v`
Expected: FAIL — `g.Run` not defined.

**Step 3: Write the implementation**

Add to `grabber.go`:

```go
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
```

**Note:** Download errors for individual emojis are logged but do not abort the entire run. The function only returns an error for fatal problems (API failure, can't create directory).

**Step 4: Run tests to verify they pass**

Run: `go test -run TestRun -v`
Expected: PASS

**Step 5: Commit**

```bash
git add grabber.go grabber_test.go
git commit -m "feat: implement Grabber.Run with alias skipping, dedup, and error handling"
```

---

## Task 6: Rewrite `main()` to use Grabber

Replace the current monolithic `main()` with a thin shell that parses flags, reads config, and delegates to `Grabber.Run()`.

**Files:**
- Modify: `slack_emoji_grabber.go` (complete rewrite)

**Step 1: Rewrite `slack_emoji_grabber.go`**

```go
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
```

**Step 2: Verify the `slack.Client` satisfies `SlackClient`**

The `slack.Client` type from `github.com/slack-go/slack` already has a `GetEmoji() (map[string]string, error)` method, so it implicitly satisfies our `SlackClient` interface. No adapter needed.

**Step 3: Verify build**

Run: `go build -o slack-emoji-grabber .`
Expected: Clean build, no errors.

**Step 4: Verify `-version` and `-h` flags**

Run: `./slack-emoji-grabber -version && ./slack-emoji-grabber -h`
Expected: Version prints `dev`, help shows flags.

**Step 5: Commit**

```bash
git add slack_emoji_grabber.go
git commit -m "refactor: rewrite main() as thin shell delegating to Grabber"
```

---

## Task 7: Add `-output` flag to Makefile build verification

Ensure the Makefile still works correctly with the refactored code.

**Files:**
- No file changes needed — verification only.

**Step 1: Run full build via Makefile**

Run: `make build`
Expected: Clean build producing `slack-emoji-grabber` binary.

**Step 2: Run all tests**

Run: `make test`
Expected: All tests pass.

**Step 3: Run lint**

Run: `make lint`
Expected: No vet errors, no fmt changes.

**Step 4: Commit (only if Makefile changes were needed)**

No commit expected for this task.

---

## Task 8: Clean up and final verification

Remove the old standalone `downloadFile` function (now a method on `Grabber`), ensure no dead code remains, and run final checks.

**Files:**
- Modify: `slack_emoji_grabber.go` — remove old `downloadFile` function if still present

**Step 1: Verify no dead code**

The old `downloadFile` function (lines 12-30 in the original) should have been removed when `main()` was rewritten in Task 6. If it's still present, delete it.

**Step 2: Run vet and tests**

Run: `go vet ./... && go test ./... -v`
Expected: No warnings, all tests pass.

**Step 3: Verify binary works end-to-end (manual)**

Run: `./slack-emoji-grabber -version`
Expected: `dev`

Run: `./slack-emoji-grabber -h`
Expected: Usage with `-output`, `-version`, `-v` flags.

**Step 4: Final commit**

```bash
git add -A
git commit -m "chore: remove dead code and finalize refactor"
```

---

## Summary of Changes

| File | Action | Purpose |
|------|--------|---------|
| `go.mod` | Modify | Fix module path (underscore → hyphen) |
| `slack_emoji_grabber.go` | Rewrite | Thin `main()` with flags, logging, version |
| `grabber.go` | Create | `SlackClient` interface, `Grabber` struct, `Run()`, `downloadFile()` |
| `grabber_test.go` | Create | Full test suite: constructor, download, run, aliases, dedup, errors |

## What This Achieves

1. **Error handling** — Every error checked, wrapped with `%w`, HTTP status validated
2. **Testability** — Interface-based Slack client, injectable HTTP client, `t.TempDir()` isolation
3. **Idiomatic Go** — `log/slog`, `flag`, `filepath.Join`, `path.Ext`, `os.MkdirAll`, `HasPrefix`
4. **No silent failures** — Download errors logged, API/filesystem errors returned
5. **HTTP timeout** — 30s default, no more hanging forever
6. **CLI flags** — `-output`, `-version`, `-v` for verbose
7. **Module path fix** — Matches repo name
