package slackclient

import (
	"context"
	"time"

	"github.com/slack-go/slack"

	"github.com/dbut2/slackgpt/pkg/models"
	"github.com/dbut2/slackgpt/pkg/slacktime"
)

type Client struct {
	slack *slack.Client
}

func New(sc *slack.Client) *Client {
	return &Client{
		slack: sc,
	}
}

func (c *Client) GetChannelMessages(channel string, from, to time.Time) ([]slack.Message, error) {
	resp, err := c.slack.GetConversationHistory(&slack.GetConversationHistoryParameters{
		ChannelID: channel,
		Inclusive: true,
		Latest:    slacktime.ParseTime(to),
		Limit:     1000,
		Oldest:    slacktime.ParseTime(from),
	})
	if err != nil {
		return nil, err
	}

	msgs := resp.Messages
	for i, j := 0, len(msgs)-1; i < len(msgs)/2; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	return msgs, nil
}

func (c *Client) GetThreadMessages(channel, threadTs string, from, to time.Time) ([]slack.Message, error) {
	msgs, _, _, err := c.slack.GetConversationReplies(&slack.GetConversationRepliesParameters{
		ChannelID: channel,
		Timestamp: threadTs,
		Inclusive: true,
		Latest:    slacktime.ParseTime(to),
		Limit:     1000,
		Oldest:    slacktime.ParseTime(from),
	})
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func (c *Client) Respond(ctx context.Context, response models.Response) error {
	_, _, err := c.slack.PostMessage(response.SlackChannel, slack.MsgOptionTS(response.SlackThreadTS), slack.MsgOptionText(response.Completion, false))
	return err
}
