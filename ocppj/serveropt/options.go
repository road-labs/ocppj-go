package serveropt

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/e-flux-platform/ocppj-go/ocppj"
)

// WithPort specifies the port to listen on.
func WithPort(port int) ocppj.ServerOption {
	return func(c *ocppj.ServerConfig) {
		c.ListenAddr = fmt.Sprintf(":%d", port)
	}
}

// WithUpgradePath specifies to upgrade path. The path can include placeholders for OCPP identities by using the
// Go 1.22+ routing enhancements, as per https://go.dev/blog/routing-enhancements
func WithUpgradePath(path string) ocppj.ServerOption {
	return func(c *ocppj.ServerConfig) {
		c.UpgradePath = path
	}
}

// WithHealthEndpoint specifies a health endpoint. This currently just exposes a single endpoint which returns a 200
// with "OK" in the response body.
func WithHealthEndpoint(path string) ocppj.ServerOption {
	return func(c *ocppj.ServerConfig) {
		c.HealthPath = path
	}
}

// WithSupportedProtocols configures which protocols are supported by the server. Any protocols not in this list will
// result in the upgrade request being rejected. The first protocol on the list to match a protocol supported by the
// client will be the agreed protocol to use.
func WithSupportedProtocols(protocols []string) ocppj.ServerOption {
	return func(c *ocppj.ServerConfig) {
		c.SupportedProtocols = protocols
	}
}

// WithReadHeaderTimeout specifies the HTTP read header timeout.
func WithReadHeaderTimeout(d time.Duration) ocppj.ServerOption {
	return func(c *ocppj.ServerConfig) {
		c.ReadHeaderTimeout = d
	}
}

// WithTLSConfig configures TLS configuration to be used by the websocket server.
func WithTLSConfig(conf *tls.Config) ocppj.ServerOption {
	return func(c *ocppj.ServerConfig) {
		c.TLSConfig = conf
	}
}

// WithUpgrader configures the websocket upgrade to use. A default upgrader is used which applies sensible defaults;
// only override it if you have a very particular reason to do so.
func WithUpgrader(upgrader ocppj.Upgrader) ocppj.ServerOption {
	return func(c *ocppj.ServerConfig) {
		c.Upgrader = upgrader
	}
}

// WithClientCallTimeout specifies the outbound call timeout to use for clients managed by the server. The default is
// 30 seconds.
func WithClientCallTimeout(timeout time.Duration) ocppj.ServerOption {
	return func(c *ocppj.ServerConfig) {
		c.ClientCallTimeout = timeout
	}
}

// WithClientRateLimiter configures a rate limiter to use when handling client messages. By default, no rate limiting
// is applied.
func WithClientRateLimiter(limiter ocppj.ClientRateLimiter) ocppj.ServerOption {
	return func(c *ocppj.ServerConfig) {
		c.ClientRateLimiter = limiter
	}
}

// WithShutdownStrategy configures a function that can be used to change the client drain strategy during shutdown. The
// default strategy is to attempt to gracefully close all clients at the same time.
func WithShutdownStrategy(strategy ocppj.ShutdownStrategy) ocppj.ServerOption {
	return func(c *ocppj.ServerConfig) {
		c.ShutdownStrategy = strategy
	}
}
