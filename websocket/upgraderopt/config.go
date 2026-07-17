package upgraderopt

import (
	"net/http"
	"time"

	"github.com/road-labs/ocppj-go/websocket"
)

// WithOriginCheck configures an origin check function to be used by the websocket upgrader.
func WithOriginCheck(fn func(*http.Request) bool) websocket.UpgraderOption {
	return func(c *websocket.UpgraderConfig) {
		c.OriginCheck = fn
	}
}

// WithClientReadTimeout specifies the read timeout to configure against upgraded websocket clients.
func WithClientReadTimeout(timeout time.Duration) websocket.UpgraderOption {
	return func(c *websocket.UpgraderConfig) {
		c.ReadTimeout = timeout
	}
}

// WithClientWriteTimeout specifies the write timeout to configure against upgraded websocket clients.
func WithClientWriteTimeout(timeout time.Duration) websocket.UpgraderOption {
	return func(c *websocket.UpgraderConfig) {
		c.WriteTimeout = timeout
	}
}

// WithClientGracefulCloseTimeout specifies the graceful close timeout to configure against upgraded websocket clients.
func WithClientGracefulCloseTimeout(timeout time.Duration) websocket.UpgraderOption {
	return func(c *websocket.UpgraderConfig) {
		c.GracefulCloseTimeout = timeout
	}
}

// WithClientPingInterval specifies the websocket ping interval to configure against upgraded websocket clients.
func WithClientPingInterval(interval time.Duration) websocket.UpgraderOption {
	return func(c *websocket.UpgraderConfig) {
		c.PingInterval = interval
	}
}
