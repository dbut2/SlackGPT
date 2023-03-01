package prompt

import (
	"fmt"
	"time"

	"github.com/slack-go/slack"

	"github.com/dbut2/slackgpt/pkg/models"
	"github.com/dbut2/slackgpt/pkg/slackclient"
	"github.com/dbut2/slackgpt/pkg/slacktime"
)

type Enhancer interface {
	Enhance(prompt models.Request) (string, error)
}

type Default struct {
	slack           *slackclient.Client
	botID           string
	historyDuration time.Duration
	historyCount    int
	maxPromptLength int
}

func NewEnhancer(sc *slackclient.Client, botID string, opts ...EnhanceOpt) Enhancer {
	e := &Default{
		slack:           sc,
		botID:           botID,
		historyDuration: time.Hour,
		historyCount:    5,
		maxPromptLength: 1000,
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

func (e *Default) Enhance(prompt models.Request) (string, error) {
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
		return "", err
	}

	if len(msgs) > e.historyCount {
		msgs = msgs[len(msgs)-1-e.historyCount:]
	}

	ss := formatMessages(msgs)

	enhanced := ""
	for _, s := range ss {
		enhanced += s + "\n"
	}

	enhanced += fmt.Sprintf("<@%s>:", e.botID)

	if len(enhanced) > e.maxPromptLength {
		enhanced = enhanced[len(enhanced)-1-e.maxPromptLength:]
	}

	return enhanced, nil
}

type EnhanceOpt func(*Default)

func WithHistoryDuration(duration time.Duration) EnhanceOpt {
	return func(e *Default) {
		e.historyDuration = duration
	}
}

func WithHistoryCount(count int) EnhanceOpt {
	return func(e *Default) {
		e.historyCount = count
	}
}

func WithMaxPromptLength(length int) EnhanceOpt {
	return func(e *Default) {
		e.maxPromptLength = length
	}
}

func formatMessage(msg slack.Message) string {
	return fmt.Sprintf("<@%s>: %s", msg.User, msg.Text)
}

func formatMessages(msgs []slack.Message) []string {
	s := make([]string, len(msgs))
	for i, msg := range msgs {
		s[i] = formatMessage(msg)
	}
	return s
}
