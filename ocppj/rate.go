package ocppj

import (
	"context"

	"github.com/e-flux-platform/ocppj-go/ocppj/message"
)

type ClientRateLimiter interface {
	Allow(ctx context.Context, client *Client, msg message.Message) bool
}
