package ocppj

import (
	"context"
	"iter"
	"time"
)

type DrainStrategy func(context.Context, iter.Seq[*Client]) error

func defaultDrainStrategy(ctx context.Context, clients iter.Seq[*Client]) error {
	for client := range clients {
		go func() {
			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()
			_ = client.GracefulClose(ctx, "server is shutting down")
		}()
	}
	return nil
}
