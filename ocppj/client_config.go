package ocppj

import (
	"crypto/tls"
	"time"
)

type ClientConfig struct {
	CallTimeout           time.Duration
	Metadata              map[string]any
	SupportedProtocols    []string
	TLSConfig             *tls.Config
	WebsocketClient       WebsocketClient
	WebsocketReadTimeout  time.Duration
	WebsocketWriteTimeout time.Duration
	WebsocketPingInterval time.Duration
	RateLimiter           ClientRateLimiter
}

type ClientOption func(*ClientConfig)

const defaultCallTimeout = time.Second * 30

func newClientConfig(opts []ClientOption) *ClientConfig {
	c := &ClientConfig{}
	for _, opt := range opts {
		opt(c)
	}
	if c.CallTimeout == 0 {
		c.CallTimeout = defaultCallTimeout
	}
	if c.Metadata == nil {
		c.Metadata = make(map[string]any)
	}
	return c
}
