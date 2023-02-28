package event

import (
	"context"
	"os"

	gogpt "github.com/sashabaranov/go-gpt3"

	"github.com/dbut2/slackgpt/internal/pubsub"
)

func Generate(ctx context.Context, m pubsub.PubSubMessage) error {
	openAIToken := os.Getenv("OPENAI_TOKEN")
	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	slackBotID := os.Getenv("SLACK_BOT_ID")
	model := os.Getenv("MODEL")

	if model == "" {
		model = gogpt.GPT3TextDavinci003
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
