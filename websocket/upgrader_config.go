package websocket

import (
	"net/http"
	"time"
)

type UpgraderConfig struct {
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	GracefulCloseTimeout time.Duration
	PingInterval         time.Duration
	OriginCheck          func(*http.Request) bool
}

type UpgraderOption func(config *UpgraderConfig)

const (
	defaultClientWriteTimeout         = time.Second * 30
	defaultClientReadTimeout          = time.Minute * 30
	defaultClientGracefulCloseTimeout = time.Second * 5
)

func newUpgraderConfig(opts []UpgraderOption) *UpgraderConfig {
	c := &UpgraderConfig{}
	for _, opt := range opts {
		opt(c)
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = defaultClientWriteTimeout
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = defaultClientReadTimeout
	}
	if c.GracefulCloseTimeout == 0 {
		c.GracefulCloseTimeout = defaultClientGracefulCloseTimeout
	}
	return c
}
