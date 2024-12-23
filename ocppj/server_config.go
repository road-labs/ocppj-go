package ocppj

import (
	"crypto/tls"
	"time"

	"github.com/e-flux-platform/ocppj-go/websocket"
)

type ServerConfig struct {
	ListenAddr         string
	UpgradePath        string
	HealthPath         string
	ClientCallTimeout  time.Duration
	ReadHeaderTimeout  time.Duration
	TLSConfig          *tls.Config
	SupportedProtocols []string
	Upgrader           Upgrader
	ShutdownStrategy   ShutdownStrategy
	ClientRateLimiter  ClientRateLimiter
}

const (
	defaultClientCallTimeout = time.Second * 30
	readHeaderTimeout        = time.Second * 30
	defaultUpgradePath       = "/"
	defaultHealthPath        = "/health"
)

type ServerOption func(*ServerConfig)

func newServerConfig(opts []ServerOption) *ServerConfig {
	c := &ServerConfig{}
	for _, opt := range opts {
		opt(c)
	}
	if c.ClientCallTimeout == 0 {
		c.ClientCallTimeout = defaultClientCallTimeout
	}
	if c.ReadHeaderTimeout == 0 {
		c.ReadHeaderTimeout = readHeaderTimeout
	}
	if c.UpgradePath == "" {
		c.UpgradePath = defaultUpgradePath
	}
	if c.HealthPath == "" {
		c.HealthPath = defaultHealthPath
	}
	if c.ShutdownStrategy == nil {
		c.ShutdownStrategy = defaultShutdownStrategy
	}
	if c.Upgrader == nil {
		c.Upgrader = websocket.NewUpgrader()
	}
	return c
}
