package prompt

import (
	"time"

	"github.com/slack-go/slack"

	"github.com/dbut2/slackgpt/pkg/models"
	"github.com/dbut2/slackgpt/pkg/slackclient"
	"github.com/dbut2/slackgpt/pkg/slacktime"
)

type MessageGetter interface {
	GetMessages(prompt models.Request) ([]Message, error)
}

type Message struct {
	Role    Role
	Message string
	Name    string
}

type Role int

const (
	UnknownRole Role = iota
	SystemRole
	AssistantRole
	UserRole
)

type defaultMessageGetter struct {
	slack           *slackclient.Client
	botID           string
	historyDuration time.Duration
	historyCount    int
	maxPromptLength int
	separator       string
}

func NewMessageGetter(sc *slackclient.Client, botID string, opts ...EnhanceOpt) MessageGetter {
	e := &defaultMessageGetter{
		slack:           sc,
		botID:           botID,
		historyDuration: time.Minute * 15,
		historyCount:    9,
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

func (e *defaultMessageGetter) GetMessages(prompt models.Request) ([]Message, error) {
	var msgs []slack.Message
	var err error
	msgTS := slacktime.ParseString(prompt.SlackMsgTS)
	switch prompt.SlackThreadTS == "" {
	case true:
		msgs, err = e.slack.GetChannelMessages(prompt.SlackChannel, msgTS.Add(-e.historyDuration), msgTS)
	case false:
		msgs, err = e.slack.GetThreadMessages(prompt.SlackChannel, prompt.SlackThreadTS, msgTS.Add(-e.historyDuration), msgTS)
	}
	if err != nil {
		return nil, err
	}

	if len(msgs) > e.historyCount {
		msgs = msgs[len(msgs)-1-e.historyCount:]
	}

	var messages []Message

	for _, msg := range msgs {
		role := UserRole
		if msg.User == e.botID {
			role = AssistantRole
		}
		messages = append(messages, Message{
			Role:    role,
			Message: msg.Text,
			Name:    msg.User,
		})
	}

	return messages, nil
}

type EnhanceOpt func(*defaultMessageGetter)

func WithHistoryDuration(duration time.Duration) EnhanceOpt {
	return func(e *defaultMessageGetter) {
		e.historyDuration = duration
	}
}

func WithHistoryCount(count int) EnhanceOpt {
	return func(e *defaultMessageGetter) {
		e.historyCount = count
	}
}

func WithMaxPromptLength(length int) EnhanceOpt {
	return func(e *defaultMessageGetter) {
		e.maxPromptLength = length
	}
}

func WithSeparator(separator string) EnhanceOpt {
	return func(e *defaultMessageGetter) {
		e.separator = separator
	}
}
