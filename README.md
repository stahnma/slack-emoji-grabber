# Grab you slack team's emojis

Do you ever wish you could grab the emojis from one slack team and keep them? Maybe you'd put them in another team? Well now you can. This simple little golang CLI will grab all the emojis and put them in `./emojis`. It also won't download an emoji if you already have it in your `./emojis` dir. It also skips aliases because seriously, you have them already.

# Setup

You need one module that isn't in the stdlib.

`go get -u github.com/nlopes/slack`

You can also simply run `make setup`

# Build

`go build slack_emoji_grabber.go`

or run

`make`

# Configure
To run the `slack_emoji_grabber` you need your `SLACK_TOKEN` environment variable set. If you don't have it set, it won't work.

# Run
`./slack_emoji_grabber`

# License
MIT
