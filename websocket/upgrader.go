package websocket

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Upgrader struct {
	upgrader             *websocket.Upgrader
	readTimeout          time.Duration
	writeTimeout         time.Duration
	gracefulCloseTimeout time.Duration
	pingInterval         time.Duration
}

func NewUpgrader(opts ...UpgraderOption) *Upgrader {
	conf := newUpgraderConfig(opts)
	delegateUpgrader := &websocket.Upgrader{}
	if conf.OriginCheck != nil {
		delegateUpgrader.CheckOrigin = conf.OriginCheck
	}
	return &Upgrader{
		upgrader:             delegateUpgrader,
		readTimeout:          conf.ReadTimeout,
		writeTimeout:         conf.WriteTimeout,
		gracefulCloseTimeout: conf.GracefulCloseTimeout,
		pingInterval:         conf.PingInterval,
	}
}

func (u *Upgrader) Upgrade(w http.ResponseWriter, req *http.Request, responseHeader http.Header) (*Client, error) {
	conn, err := u.upgrader.Upgrade(w, req, responseHeader)
	if err != nil {
		return nil, err
	}
	return newClient(
		conn,
		req.Host,
		u.readTimeout,
		u.writeTimeout,
		u.gracefulCloseTimeout,
		u.pingInterval,
	), nil
}
