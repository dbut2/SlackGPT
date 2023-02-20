package event

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	gogpt "github.com/sashabaranov/go-gpt3"
	"github.com/slack-go/slack"
	"google.golang.org/protobuf/proto"

	"github.com/dbut2/slackgpt/proto/pkg"
)

// PubSubMessage is the payload of a Pub/Sub event. Please refer to the docs for
// additional information regarding Pub/Sub events.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

func GenerateFromPubSub(ctx context.Context, m PubSubMessage) error {
	req := new(pkg.Request)
	err := proto.Unmarshal(m.Data, req)
	if err != nil {
		return err
	}

	return processPrompt(req.Channel, req.ThreadTimestamp, req.Prompt)
}

func processPrompt(channel string, thread_ts string, prompt string) error {
	c := gogpt.NewClient(os.Getenv("OPENAI_TOKEN"))

	prompt = strings.ReplaceAll(prompt, os.Getenv("SLACK_BOT_ID"), "SlackGPT")

	req := gogpt.CompletionRequest{
		Model:     gogpt.GPT3TextDavinci003,
		Prompt:    prompt,
		MaxTokens: 1000,
	}

	resp, err := c.CreateCompletion(context.Background(), req)

	bo := backoff.ExponentialBackOff{
		InitialInterval: time.Second,
		Multiplier:      1.5,
	}

	for err != nil {
		if reqErr, ok := err.(*gogpt.RequestError); ok {
			if reqErr.StatusCode == http.StatusTooManyRequests {
				time.Sleep(bo.NextBackOff())
				resp, err = c.CreateCompletion(context.Background(), req)
			} else {
				return err
			}
		} else {
			return err
		}
	}

	if resp.Choices[0].Text == "" {
		return nil
	}

	sc := slack.New(os.Getenv("SLACK_BOT_TOKEN"))

	_, _, err = sc.PostMessage(channel, slack.MsgOptionTS(thread_ts), slack.MsgOptionText(resp.Choices[0].Text, false))
	if err != nil {
		return err
	}

	return nil
}
