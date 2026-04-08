# slack-emoji-grabber

A simple CLI tool to download and archive your Slack workspace's custom emojis. It saves them to a local directory, skips aliases, and avoids re-downloading emojis you already have.

## Install

```bash
go install github.com/stahnma/slack-emoji-grabber@latest
```

Or build from source:

```bash
make build
```

## Configure

You need a **Slack Bot User OAuth Token** with the `emoji:read` scope. Find it under "OAuth & Permissions" in your [Slack App settings](https://api.slack.com/apps).

```bash
export SLACK_TOKEN="xoxb-1234567890-0987654321-AbCdEfGhIjKlMnOpQrStUvWxYz"
```

## Usage

```bash
slack-emoji-grabber [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-output` | `emojis` | Directory to save emojis |
| `-v` | `false` | Enable verbose/debug logging |
| `-version` | | Print version and exit |

### Examples

```bash
# Download all emojis to ./emojis
slack-emoji-grabber

# Download to a custom directory
slack-emoji-grabber -output /tmp/my-emojis

# See what's happening
slack-emoji-grabber -v
```

## Development

```bash
make help      # Show available targets
make build     # Build the binary
make test      # Run tests
make lint      # Run vet + gofmt
```

## License

MIT
