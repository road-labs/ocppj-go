package ocppj

import (
	"context"

	"github.com/road-labs/ocppj-go/ocppj/message"
)

type ClientRateLimiter interface {
	Allow(ctx context.Context, client *Client, msg message.Message) bool
}
