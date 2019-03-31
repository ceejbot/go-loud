# GO-LOUD

GO LOUD OR GO HOME. LOUDBOT IS A SLACK BOT THAT SHOUTS AT YOU IF YOU SHOUT AT IT. SHOUTING IS CATHARTIC.

Configuration is injected from environment variables. A backing redis is required to remember what was shouted across runs.

## RUNNING

1. Create an application in Slack. Name it LOUDBOT or something similar.
2. Add a bot user to your app and snag its API token. Invite the bot user to a couple of channels
3. Set up a redis somewhere.
4. Create an env file or copy the example: `cp env.example .env`. Edit it so it points to your redis and uses your slack token:

```
REDIS_ADDRESS=localhost:6379
SLACK_TOKEN=<your slack api token>
WELCOME_CHANNEL=general # optional; loudie will toast here
REDIS_PREFIX=LOUDBOT # defaults to `LB`; namespace for redis keys
```

You can skip this step if you have another way to provide the required env vars.


5. Now you'll want to seed the yell database. If you like, look at `SEEDS` and add or remove yells to suit your team.

```sh
cd cmd/seedlouds
go build
./seedlouds # db is now seeded
```

6. Run LOUDBOT.

```sh
cd ../..
go build
./go-loud # loudie is now running
```

There's a dockerfile if you prefer not to trust your operating system.

## TODO

- Read list of custom emoji & react to messages sometimes.
- More Malcolm Tucker, but behind a feature flag.
- ✔︎ Tool to dump redis db to a text file.
- More random jokes.

## LICENSE

ISC.
