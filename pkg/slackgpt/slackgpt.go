package slackgpt

import (
	"context"

	"github.com/dbut2/slackgpt/pkg/models"
)

type Sender interface {
	Send(ctx context.Context, request models.Request) error
}

type Responder interface {
	Respond(ctx context.Context, response models.Response) error
}
