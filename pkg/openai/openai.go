package openai

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sashabaranov/go-openai"

	"github.com/dbut2/slackgpt/pkg/models"
	"github.com/dbut2/slackgpt/pkg/prompt"
	"github.com/dbut2/slackgpt/pkg/slackgpt"
)

type Client struct {
	openai    *openai.Client
	mg        prompt.MessageGetter
	responder slackgpt.Responder
	model     string
}

func New(token string, mg prompt.MessageGetter, responder slackgpt.Responder, model string, opts ...ClientOption) *Client {
	client := openai.NewClient(token)

	if model == "" {
		model = openai.GPT3Dot5Turbo
	}

	c := &Client{
		openai:    client,
		mg:        mg,
		responder: responder,
		model:     model,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type ClientOption func(client *Client)

func (c *Client) Send(ctx context.Context, req models.Request) error {
	msgs, err := c.mg.GetMessages(req)
	if err != nil {
		return err
	}

	ms := mapMessages(msgs)

	for msgsLength(ms) > 3072*4 {
		ms = ms[1:]
	}

	r := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		Messages:  ms,
		MaxTokens: 1024,
		User:      req.User,
	}

	resp, err := c.openai.CreateChatCompletion(ctx, r)

	bo := backoff.ExponentialBackOff{
		InitialInterval: time.Second,
		Multiplier:      1.5,
	}
	for err != nil {
		if reqErr, ok := err.(*openai.RequestError); ok {
			if reqErr.StatusCode == http.StatusTooManyRequests {
				nbo := bo.NextBackOff()
				log.Printf("OpenAI rate limit: %s", nbo.String())
				time.Sleep(nbo)
				resp, err = c.openai.CreateChatCompletion(context.Background(), r)
			} else {
				return err
			}
		} else {
			return err
		}
	}

	if len(resp.Choices) < 1 {
		return errors.New("no completions returned")
	}

	if resp.Choices[0].Message.Content == "" {
		return errors.New("empty completion text")
	}

	return c.responder.Respond(context.Background(), models.Response{
		Completion:    resp.Choices[0].Message.Content,
		Timestamp:     time.Now(),
		SlackChannel:  req.SlackChannel,
		SlackThreadTS: req.SlackThreadTS,
	})
}

var roleMap = map[prompt.Role]string{
	prompt.SystemRole:    openai.ChatMessageRoleSystem,
	prompt.AssistantRole: openai.ChatMessageRoleAssistant,
	prompt.UserRole:      openai.ChatMessageRoleUser,
}

func mapMessages(msgs []prompt.Message) []openai.ChatCompletionMessage {
	var ms []openai.ChatCompletionMessage

	for _, msg := range msgs {
		ms = append(ms, openai.ChatCompletionMessage{
			Role:    roleMap[msg.Role],
			Content: msg.Message,
			Name:    msg.Name,
		})
	}

	return ms
}

func msgsLength(msgs []openai.ChatCompletionMessage) int {
	c := 0

	for _, msg := range msgs {
		c += len(msg.Content)
	}

	return c
}
