package ocppj

import (
	"context"
	"time"
)

type DrainStrategy func(context.Context, []*Client) error

func defaultDrainStrategy(ctx context.Context, clients []*Client) error {
	for _, client := range clients {
		go func() {
			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()
			_ = client.GracefulClose(ctx, "server is shutting down")
		}()
	}
	return nil
}
