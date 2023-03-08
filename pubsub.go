package event

import (
	"context"
	"os"

	"github.com/dbut2/slackgpt/internal/pubsub"
)

func PubSubGenerate(ctx context.Context, m pubsub.PubSubMessage) error {
	openAIToken := os.Getenv("OPENAI_TOKEN")
	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	model := os.Getenv("MODEL")

	ps, err := pubsub.New(pubsub.Config{
		OpenAIToken:   openAIToken,
		SlackBotToken: slackBotToken,
		Model:         model,
	})
	if err != nil {
		return err
	}

	return ps.GenerateFromPubSub(ctx, m)
}
