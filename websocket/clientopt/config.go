package clientopt

import (
	"crypto/tls"
	"time"

	"github.com/road-labs/ocppj-go/websocket"
)

// WithSubprotocols configures the subprotocols that will be offered by the client.
func WithSubprotocols(subprotocols []string) websocket.ClientOption {
	return func(c *websocket.ClientConfig) {
		c.Subprotocols = subprotocols
	}
}

// WithReadTimeout can be used to configure the read timeout on the websocket connection. If set, this is the amount of
// time that is allowed to elapse without any data being received from the other party. Once the timeout is reached, the
// connection is closed. The default is 0, which means no timeout is applied.
func WithReadTimeout(timeout time.Duration) websocket.ClientOption {
	return func(c *websocket.ClientConfig) {
		c.ReadTimeout = timeout
	}
}

// WithWriteTimeout can be used to configure a write deadline on the underlying websocket connection. This applies
// per-write. If the timeout is hit, the connection is closed. The default is 30 seconds.
func WithWriteTimeout(timeout time.Duration) websocket.ClientOption {
	return func(c *websocket.ClientConfig) {
		c.WriteTimeout = timeout
	}
}

// WithGracefulCloseTimeout configures how long the client implementation should wait for a reply to a CLOSE frame
// sent during a call to GracefulClose().
func WithGracefulCloseTimeout(timeout time.Duration) websocket.ClientOption {
	return func(c *websocket.ClientConfig) {
		c.GracefulCloseTimeout = timeout
	}
}

// WithTLSConfig configures the TLS configuration to use when sending the websocket upgrade request.
func WithTLSConfig(tlsConfig *tls.Config) websocket.ClientOption {
	return func(c *websocket.ClientConfig) {
		c.TLSConfig = tlsConfig
	}
}

// WithPingInterval specifies how often the client should send a websocket ping. The default is 0, which means no pings
// are sent.
func WithPingInterval(interval time.Duration) websocket.ClientOption {
	return func(c *websocket.ClientConfig) {
		c.PingInterval = interval
	}
}
