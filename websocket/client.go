package websocket

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client is a thin abstraction over a gorilla/websocket connection. The main features this provides over the basic
// gorilla implementation is:
//
// - concurrency safe by default
// - ping handling enabled by default
// - graceful websocket close using websocket close frames
// - read and write timeouts automatically applied
type Client struct {
	id                     string
	remoteHost             string
	connMux                sync.Mutex
	conn                   *websocket.Conn
	readTimeout            time.Duration
	writeTimeout           time.Duration
	gracefulCloseTimeout   time.Duration
	gracefulCloseInitiated atomic.Bool
	gracefulCloseDone      chan struct{}
	closeReceived          chan struct{}
	cancel                 context.CancelFunc
}

// Open creates a connection to a websocket server
func Open(ctx context.Context, targetURL string, opts ...ClientOption) (*Client, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	conf := newClientConfig(opts)
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 30 * time.Second,
		Subprotocols:     conf.Subprotocols,
		TLSClientConfig:  conf.TLSConfig,
	}

	conn, _, err := dialer.DialContext(ctx, targetURL, nil)
	if err != nil {
		return nil, err
	}

	return newClient(conn, u.Host, conf.ReadTimeout, conf.WriteTimeout, conf.GracefulCloseTimeout, conf.PingInterval), nil
}

func newClient(
	conn *websocket.Conn,
	remoteHost string,
	readTimeout time.Duration,
	writeTimeout time.Duration,
	gracefulCloseTimeout time.Duration,
	pingInterval time.Duration,
) *Client {
	pingCtx, cancel := context.WithCancel(context.Background())
	c := &Client{
		id:                   uuid.NewString(),
		conn:                 conn,
		remoteHost:           remoteHost,
		readTimeout:          readTimeout,
		writeTimeout:         writeTimeout,
		gracefulCloseTimeout: gracefulCloseTimeout,
		gracefulCloseDone:    make(chan struct{}),
		closeReceived:        make(chan struct{}),
		cancel:               cancel,
	}
	c.conn.SetPingHandler(c.pingHandler)
	c.conn.SetCloseHandler(c.closeHandler)
	if pingInterval != 0 {
		go func() {
			c.sendPings(pingCtx, pingInterval)
		}()
	}
	return c
}

// ID returns a unique ID for this connection
func (c *Client) ID() string {
	return c.id
}

// Read consumes a message from the other party
func (c *Client) Read() (Message, error) {
	if c.readTimeout != 0 {
		if err := c.conn.SetReadDeadline(time.Now().Add(c.readTimeout)); err != nil {
			return Message{}, fmt.Errorf("failed to set read deadline: %w", err)
		}
	}
	messageType, data, err := c.conn.ReadMessage()
	if err != nil {
		return Message{}, fmt.Errorf("failed to read message: %w", err)
	}
	return Message{
		Type: MessageType(messageType),
		Data: data,
	}, nil
}

// Write sends a message to the other party
func (c *Client) Write(msg Message) error {
	c.connMux.Lock()
	defer c.connMux.Unlock()

	if msg.Type.IsControl() {
		return c.conn.WriteControl(int(msg.Type), msg.Data, time.Now().Add(c.writeTimeout))
	}

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); err != nil {
		return err
	}
	return c.conn.WriteMessage(int(msg.Type), msg.Data)
}

// Subprotocol returns the websocket subprotocol
func (c *Client) Subprotocol() string {
	return c.conn.Subprotocol()
}

// LocalAddr returns the local address of the client
func (c *Client) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteHost returns the host the client is connected to
func (c *Client) RemoteHost() string {
	return c.remoteHost
}

// GracefulClose attempts to gracefully close the websocket connection by starting a websocket close handshake. It is
// recommended that the supplied context has some timeout associated with it, as misbehaving clients may not respond
// to the close attempt.
func (c *Client) GracefulClose(ctx context.Context, text string) error {
	if !c.gracefulCloseInitiated.CompareAndSwap(false, true) {
		return errors.New("graceful close can only be initiated once")
	}

	ctx, cancel := context.WithTimeout(ctx, c.gracefulCloseTimeout)
	defer cancel()

	// Used to signal that we're not listening for any close replies
	defer close(c.gracefulCloseDone)

	// Send close message to other party
	msg := Message{
		Type: MessageTypeClose,
		Data: websocket.FormatCloseMessage(websocket.CloseGoingAway, text),
	}
	if err := c.Write(msg); err != nil && !errors.Is(err, websocket.ErrCloseSent) {
		// Failed to write the close frame, so we'll just shut down the connection
		return c.Close()
	}

	select {
	case <-ctx.Done():
		err := c.Close() // attempt to close anyway
		return errors.Join(ctx.Err(), err)
	case <-c.closeReceived:
		// Received a close response from the other party, so we can safely terminate the connection
		return c.Close()
	}
}

func (c *Client) sendPings(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.Write(Message{Type: MessageTypePing}); err != nil {
				slog.Error("failed to send ping", slog.Any("error", err))
			}
		}
	}
}

func (c *Client) pingHandler(data string) error {
	return c.Write(Message{
		Type: MessageTypePong,
		Data: []byte(data),
	})
}

func (c *Client) closeHandler(code int, text string) error {
	if c.gracefulCloseInitiated.Load() {
		select {
		case c.closeReceived <- struct{}{}:
		case <-c.gracefulCloseDone: // Nobody listening
		}
		return nil
	}
	return c.Write(Message{
		Type: MessageTypeClose,
		Data: websocket.FormatCloseMessage(code, text),
	})
}

// Close closes the underlying connection
func (c *Client) Close() error {
	c.cancel()
	return c.conn.Close()
}
