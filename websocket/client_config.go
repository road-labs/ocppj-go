package websocket

import (
	"crypto/tls"
	"time"
)

type ClientConfig struct {
	Subprotocols         []string
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	GracefulCloseTimeout time.Duration
	TLSConfig            *tls.Config
	PingInterval         time.Duration
}

type ClientOption func(*ClientConfig)

const (
	defaultWriteTimeout         = time.Second * 30
	defaultGracefulCloseTimeout = time.Second * 5
)

func newClientConfig(opts []ClientOption) *ClientConfig {
	c := &ClientConfig{}
	for _, opt := range opts {
		opt(c)
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = defaultWriteTimeout
	}
	if c.GracefulCloseTimeout == 0 {
		c.GracefulCloseTimeout = defaultGracefulCloseTimeout
	}
	return c
}
