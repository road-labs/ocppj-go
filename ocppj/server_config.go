package ocppj

import (
	"crypto/tls"
	"log/slog"
	"time"

	"github.com/road-labs/ocppj-go/websocket"
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
	DrainStrategy      DrainStrategy
	ClientRateLimiter  ClientRateLimiter
	Logger             *slog.Logger
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
	if c.DrainStrategy == nil {
		c.DrainStrategy = defaultDrainStrategy
	}
	if c.Upgrader == nil {
		c.Upgrader = websocket.NewUpgrader()
	}
	if c.Logger == nil {
		c.Logger = slog.New(slog.DiscardHandler)
	}
	return c
}
