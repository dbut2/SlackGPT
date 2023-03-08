# SlackGPT

Slack bot for generating chats or text completions using OpenAI's range of [models](https://platform.openai.com/docs/models).

SlackGPT is built on top of [Slack Bot Users](https://api.slack.com/bot-users) for bot interactions and messages, [OpenAI APIs](https://platform.openai.com/docs/api-reference/introduction) for AI text completion, and GCP for sewing everything together, notably [Cloud Functions](https://cloud.google.com/functions) and [Pub/Sub](https://cloud.google.com/pubsub/docs/overview).

SlackGPT can be direct messaged privately if you need a listening ear, or added to a channel for a team's worth of fun.

![](image.png)

## Usage

There's 2 ways of running SlackGPT. Firstly through GCP using Cloud Functions and PubSub, this is the preferred method as Cloud Functions can manage scaling for you and be generally cheaper, and PubSub also helps handle error correction and retry for generation. The other option is to run in socket mode which is also useful for testing purposes.

Either way, you'll need to ensure you have the following resources:
- OpenAI:
  - [API Key](https://platform.openai.com/account/api-keys)

- Slack:
  - [Bot Token](https://api.slack.com/authentication/token-types)
  - [App Token](https://api.slack.com/authentication/token-types) (if using socket mode)

- GCP (if using pubsub mode):
  - [Project ID](https://cloud.google.com/resource-manager/docs/creating-managing-projects)
  - [PubSub Topic](https://cloud.google.com/pubsub/docs/create-topic)

### PubSub mode

Create generating function from [`pubsub.go;PubSubGenerate`](pubsub.go#L10)

This function should be triggered from pubsub topic push and should have the following env vars set:
```
OPENAI_TOKEN=(OpenAI API Key)
SLACK_BOT_TOKEN=(Slack Bot token)
MODEL=(Optional. Model for text generation. Defaults to "gpt-3.5-turbo")
```

Create slack event receiver function from [`event.go;SlackEvent`](event.go#L11)

This function should be triggered from HTTP, provide the trigger URL to Slack event subscriptions and should have the following env vars set:
```
SLACK_SINGING_SECRET=(Signing Secret from Slack app)
PROJECT_ID=(GCP Project ID)
PUBSUB_TOPIC=(GCP PubSub Topic)
```

### Socket mode

Set the following env vars:
```
export OPENAI_TOKEN=(Slack API Key)
export SLACK_APP_TOKEN=(Slack App Token)
export SLACK_BOT_TOKEN=(Slack Bot Token)
export MODEL=(Optional. Model for text generation. Defaults to "gpt-3.5-turbo")
```

Run socket mode application:
```
go run cmd/socket/main.go
```

Enable socket mode in Slack app settings
