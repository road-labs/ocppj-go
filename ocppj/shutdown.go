package ocppj

import (
	"context"
	"time"
)

type ShutdownStrategy func(context.Context, []*Client) error

func defaultShutdownStrategy(ctx context.Context, clients []*Client) error {
	for _, client := range clients {
		go func() {
			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()
			_ = client.GracefulClose(ctx, "server is shutting down")
		}()
	}
	return nil
}
