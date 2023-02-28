package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/dbut2/slackgpt/pkg/models"
	"github.com/dbut2/slackgpt/proto/pkg"
)

type Sender struct {
	topic *pubsub.Topic
}

func New(topic *pubsub.Topic) *Sender {
	return &Sender{
		topic: topic,
	}
}

func (c *Sender) Send(ctx context.Context, req models.Request) error {
	r := &pkg.Request{
		Prompt:               req.Prompt,
		User:                 req.User,
		Timestamp:            timestamppb.New(req.Timestamp),
		SlackChannel:         req.SlackChannel,
		SlackThreadTimestamp: req.SlackThreadTS,
		SlackMsgTimestamp:    req.SlackMsgTS,
	}

	b, err := proto.Marshal(r)
	if err != nil {
		return err
	}
	msg := &pubsub.Message{
		Data: b,
	}

	res := c.topic.Publish(ctx, msg)
	_, err = res.Get(ctx)
	if err != nil {
		return err
	}

	return nil
}
