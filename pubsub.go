package event

import (
	"context"
	"os"

	"github.com/dbut2/slackgpt/internal/pubsub"
	"github.com/dbut2/slackgpt/pkg/openai"
)

func PubSubGenerate(ctx context.Context, m pubsub.PubSubMessage) error {
	openAIToken := os.Getenv("OPENAI_TOKEN")
	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	slackBotID := os.Getenv("SLACK_BOT_ID")
	model := os.Getenv("MODEL")

	if model == "" {
		model = openai.TextDavinci003
	}

	ps, err := pubsub.New(pubsub.Config{
		OpenAIToken:   openAIToken,
		SlackBotToken: slackBotToken,
		SlackBotID:    slackBotID,
		Model:         model,
	})
	if err != nil {
		return err
	}

	return ps.GenerateFromPubSub(ctx, m)
}
