# GO-LOUD

GO LOUD OR GO HOME. LOUDBOT IS A SLACK BOT THAT SHOUTS AT YOU IF YOU SHOUT AT IT. SHOUTING IS CATHARTIC.

## RUNNING

Configuration is injected from environment variables. A backing redis is required to remember what was shouted across runs.

1. Create an application in Slack. Name it LOUDBOT or something similar.
2. Add a bot user to your app and snag its API token.
3. Set up a redis somewhere.
4. Create an env file or copy the example: `cp env.example .env`. Edit it so it points to your redis and uses your slack token:

```
REDIS_ADDRESS=localhost:6379
SLACK_TOKEN=<your slack api token>
WELCOME_CHANNEL=general # optional; loudie will toast here
REDIS_KEY=LOUDBOT_YELLS # optional; namespace for redis keys
```

You can skip this step if you have another way to provide the required env vars.


5. Now you'll want to seed the yell database and run the bot:

```sh
cd cmd/seedlouds
go build
./seedlouds # db is now seeded
cd cmd/go-loud
go build
./go-loud # loudie is now running
```

## LICENSE

ISC.
