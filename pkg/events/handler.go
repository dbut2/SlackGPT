package events

import (
	"context"
	"time"

	"github.com/slack-go/slack/slackevents"

	"github.com/dbut2/slackgpt/pkg/models"
	"github.com/dbut2/slackgpt/pkg/slackgpt"
)

type Handler interface {
	HandleAppMentionEvent(a *slackevents.AppMentionEvent) error
	HandleMessageEvent(a *slackevents.MessageEvent) error
}

type DefaultHandler struct {
	sender slackgpt.Sender
}

func New(sender slackgpt.Sender) Handler {
	return &DefaultHandler{
		sender: sender,
	}
}

func (h *DefaultHandler) HandleAppMentionEvent(a *slackevents.AppMentionEvent) error {
	return h.sender.Send(context.Background(), models.Request{
		Prompt:        a.Text,
		User:          a.User,
		Timestamp:     time.Now(),
		SlackChannel:  a.Channel,
		SlackThreadTS: a.ThreadTimeStamp,
		SlackMsgTS:    a.TimeStamp,
	})
}

func (h *DefaultHandler) HandleMessageEvent(a *slackevents.MessageEvent) error {
	if a.BotID != "" {
		return nil
	}

	return h.sender.Send(context.Background(), models.Request{
		Prompt:        a.Text,
		User:          a.User,
		Timestamp:     time.Now(),
		SlackChannel:  a.Channel,
		SlackThreadTS: a.ThreadTimeStamp,
		SlackMsgTS:    a.TimeStamp,
	})
}
