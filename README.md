# twitchChangelog

A small go script scrapping Twitch changelog and warn on a Discord channel (using webhook) if an update is detected

## Build

```
$ go build twitchChangelog.go
```

## Launch

```
./twitchChangelog -discord=https://discord.com/api/webhooks/xxxxxx
```

### Discord webhook
 - https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks
