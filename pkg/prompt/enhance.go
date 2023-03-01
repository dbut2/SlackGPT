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
	separator       string
}

func NewEnhancer(sc *slackclient.Client, botID string, opts ...EnhanceOpt) Enhancer {
	e := &Default{
		slack:           sc,
		botID:           botID,
		historyDuration: time.Minute * 15,
		historyCount:    9,
		maxPromptLength: 1000,
		separator:       "\n\n---\n\n",
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

	ss, err := e.formatMessages(msgs)
	if err != nil {
		return "", err
	}

	bot, err := e.slack.GetUserInfo(e.botID)
	if err != nil {
		return "", err
	}
	preamble := fmt.Sprintf("Welcome to the conversation. Your name is %s and you are a Slack Bot chat bot. Your username is <@%s>.\n\n", bot.Profile.FirstName, bot.ID)

	enhanced := ""
	for _, s := range ss {
		enhanced += s + e.separator
	}
	enhanced += fmt.Sprintf("[%s <@%s>]: ", bot.Name, bot.ID)

	if len(enhanced)+len(preamble) > e.maxPromptLength {
		enhanced = enhanced[len(enhanced)+len(preamble)-1-e.maxPromptLength:]
	}

	enhanced = preamble + enhanced

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

func WithSeparator(separator string) EnhanceOpt {
	return func(e *Default) {
		e.separator = separator
	}
}

func (e *Default) formatMessage(msg slack.Message) (string, error) {
	user, err := e.slack.GetUserInfo(msg.User)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("[\"%s\" <@%s>]: %s", user.Profile.FirstName, user.ID, msg.Text), nil
}

func (e *Default) formatMessages(msgs []slack.Message) ([]string, error) {
	s := make([]string, len(msgs))
	var err error
	for i, msg := range msgs {
		s[i], err = e.formatMessage(msg)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}
