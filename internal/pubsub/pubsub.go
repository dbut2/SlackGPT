package pubsub

import (
	"context"

	"github.com/slack-go/slack"
	"google.golang.org/protobuf/proto"

	"github.com/dbut2/slackgpt/pkg/models"
	"github.com/dbut2/slackgpt/pkg/openai"
	"github.com/dbut2/slackgpt/pkg/prompt"
	"github.com/dbut2/slackgpt/pkg/slackclient"
	"github.com/dbut2/slackgpt/pkg/slackgpt"
	"github.com/dbut2/slackgpt/pkg/slacktime"
	"github.com/dbut2/slackgpt/proto/pkg"
)

type Config struct {
	OpenAIToken   string
	SlackBotToken string
	SlackBotID    string
	Model         string
}

type Pubsub struct {
	sender slackgpt.Sender
}

func New(config Config) (*Pubsub, error) {
	sc := slackclient.New(slack.New(config.SlackBotToken))
	e := prompt.NewMessageGetter(sc, config.SlackBotID)

	sender := openai.New(config.OpenAIToken, e, sc, config.Model)

	return &Pubsub{
		sender: sender,
	}, nil
}

type PubSubMessage struct {
	Data []byte `json:"data"`
}

func (p *Pubsub) GenerateFromPubSub(ctx context.Context, m PubSubMessage) error {
	req := new(pkg.Request)
	err := proto.Unmarshal(m.Data, req)
	if err != nil {
		return err
	}

	return p.sender.Send(ctx, models.Request{
		Prompt:        req.Prompt,
		User:          req.User,
		Timestamp:     slacktime.ParseString(req.Timestamp),
		SlackChannel:  req.SlackChannel,
		SlackThreadTS: req.SlackThreadTimestamp,
		SlackMsgTS:    req.SlackMsgTimestamp,
	})
}
