package clientopt

import (
	"crypto/tls"
	"log/slog"
	"time"

	"github.com/road-labs/ocppj-go/ocppj"
)

// WithCallTimeout specifies how long to wait for a reply to an outbound call before it is treated as lost. If not set,
// a default value of 30 seconds is used.
func WithCallTimeout(timeout time.Duration) ocppj.ClientOption {
	return func(c *ocppj.ClientConfig) {
		c.CallTimeout = timeout
	}
}

// WithMetadata can be used to supply any initial metadata to be assigned to the client instance.
func WithMetadata(metadata map[string]any) ocppj.ClientOption {
	return func(c *ocppj.ClientConfig) {
		c.Metadata = metadata
	}
}

// WithSupportedProtocols specified which OCPP protocols the client instance supports. The values should match the
// protocol strings defined in the OCPP specifications, for example:
// - ocpp1.5
// - ocpp1.6
// - ocpp2.0.1
// - ocpp2.1
func WithSupportedProtocols(protocols []string) ocppj.ClientOption {
	return func(c *ocppj.ClientConfig) {
		c.SupportedProtocols = protocols
	}
}

// WithTLSConfig can be used to specify the TLS configuration to use when opening a websocket connection.
func WithTLSConfig(tlsConfig *tls.Config) ocppj.ClientOption {
	return func(c *ocppj.ClientConfig) {
		c.TLSConfig = tlsConfig
	}
}

// WithWebsocketReadTimeout can be used to configure the read timeout on the websocket connection. If set, this is the
// amount of time that is allowed to elapse without any data being received from the other party. Once the timeout is
// reached, the connection is closed. The default is 0, which means no timeout is applied.
func WithWebsocketReadTimeout(timeout time.Duration) ocppj.ClientOption {
	return func(c *ocppj.ClientConfig) {
		c.WebsocketReadTimeout = timeout
	}
}

// WithWebsocketWriteTimeout can be used to configure a write deadline on the underlying websocket connection. This
// applies per-write. If the timeout is hit, the connection is closed. The default is 0, which means the default set by
// the websocket package is used (currently 30 seconds)
func WithWebsocketWriteTimeout(timeout time.Duration) ocppj.ClientOption {
	return func(c *ocppj.ClientConfig) {
		c.WebsocketWriteTimeout = timeout
	}
}

// WithWebsocketPingInterval specifies how often the client should send a websocket ping. This can often be a
// requirement for maintaining a stable connection to an OCPP backend. The default is 0, which means no pings are sent.
func WithWebsocketPingInterval(interval time.Duration) ocppj.ClientOption {
	return func(c *ocppj.ClientConfig) {
		c.WebsocketPingInterval = interval
	}
}

// WithRateLimiter can be used to configure a rate limiter implementation. By default there is no underlying rate
// limiting applied.
func WithRateLimiter(rl ocppj.ClientRateLimiter) ocppj.ClientOption {
	return func(c *ocppj.ClientConfig) {
		c.RateLimiter = rl
	}
}

// WithLogger configures the logger to use for the client instance. If not set, there is no log output.
func WithLogger(logger *slog.Logger) ocppj.ClientOption {
	return func(c *ocppj.ClientConfig) {
		c.Logger = logger
	}
}

// WithWebsocketClient can be used to force an existing websocket client be used, instead of opening a new connection.
// This is primarily for test cases to use.
func WithWebsocketClient(wsClient ocppj.WebsocketClient) ocppj.ClientOption {
	return func(c *ocppj.ClientConfig) {
		c.WebsocketClient = wsClient
	}
}
