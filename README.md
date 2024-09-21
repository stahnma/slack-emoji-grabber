# Grab you slack team's emojis

Do you ever wish you could grab the emojis from one slack team and keep them? Maybe you'd put them in another team? Well now you can. This simple little golang CLI will grab all the emojis and put them in `./emojis`. It also won't download an emoji if you already have it in your `./emojis` dir. It also skips aliases because seriously, you have them already.

# Setup

You can also simply run `make setup`

# Build

`go build slack_emoji_grabber.go`

or run

`make`

# Configure
To run the `slack_emoji_grabber`, you need to set your **Slack Bot User OAuth Token** as `SLACK_TOKEN` in an environment variable set. The Slack Bot User OAuth Token can be found under "OAuth & Permissions" in Slack API: Applications settings under a configured & connected Slack Application for the workspace. 

Example: `export SLACK_TOKEN="xoxb-1234567890-0987654321-AbCdEfGhIjKlMnOpQrStUvWxYz"`


# Run
`./slack_emoji_grabber`

# License
MIT

