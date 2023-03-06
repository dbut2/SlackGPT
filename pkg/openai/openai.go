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

const TextDavinci003 = openai.GPT3TextDavinci003

type Client struct {
	openai    *openai.Client
	enhancer  prompt.Enhancer
	responder slackgpt.Responder
	model     string
	separator string
}

func New(token string, enhancer prompt.Enhancer, responder slackgpt.Responder, model string, opts ...ClientOption) *Client {
	client := openai.NewClient(token)

	if model == "" {
		model = openai.GPT3TextDavinci003
	}

	c := &Client{
		openai:    client,
		enhancer:  enhancer,
		responder: responder,
		model:     model,
		separator: "\n\n---\n\n",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type ClientOption func(client *Client)

func WithSeparator(separator string) ClientOption {
	return func(c *Client) {
		c.separator = separator
	}
}

func (c *Client) Send(ctx context.Context, req models.Request) error {
	enhanced, err := c.enhancer.Enhance(req)
	if err != nil {
		return err
	}

	r := openai.CompletionRequest{
		Model:     c.model,
		Prompt:    enhanced,
		MaxTokens: 1000,
		Stop:      []string{c.separator},
		User:      req.User,
	}

	resp, err := c.openai.CreateCompletion(ctx, r)

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
				resp, err = c.openai.CreateCompletion(context.Background(), r)
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

	if resp.Choices[0].Text == "" {
		return errors.New("empty completion text")
	}

	return c.responder.Respond(context.Background(), models.Response{
		Completion:    resp.Choices[0].Text,
		Timestamp:     time.Now(),
		SlackChannel:  req.SlackChannel,
		SlackThreadTS: req.SlackThreadTS,
	})
}
