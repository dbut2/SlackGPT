package events

import (
	"context"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	"github.com/dbut2/slackgpt/pkg/models"
	"github.com/dbut2/slackgpt/pkg/slackgpt"
)

type Handler interface {
	HandleAppMentionEvent(a *slackevents.AppMentionEvent) error
	HandleMessageEvent(a *slackevents.MessageEvent) error
}

type DefaultHandler struct {
	slack  *slack.Client
	sender slackgpt.Sender
}

func New(slack *slack.Client, sender slackgpt.Sender) Handler {
	return &DefaultHandler{
		slack:  slack,
		sender: sender,
	}
}

func (h *DefaultHandler) HandleAppMentionEvent(a *slackevents.AppMentionEvent) error {
	user, err := h.slack.GetUserInfo(a.User)
	if err != nil {
		return err
	}

	return h.sender.Send(context.Background(), models.Request{
		Prompt:        a.Text,
		User:          user.ID,
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

	user, err := h.slack.GetUserInfo(a.User)
	if err != nil {
		return err
	}

	return h.sender.Send(context.Background(), models.Request{
		Prompt:        a.Text,
		User:          user.ID,
		Timestamp:     time.Now(),
		SlackChannel:  a.Channel,
		SlackThreadTS: a.ThreadTimeStamp,
		SlackMsgTS:    a.TimeStamp,
	})
}
