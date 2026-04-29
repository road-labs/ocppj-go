package ocppj

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/e-flux-platform/ocppj-go/ocppj/message"
	"github.com/e-flux-platform/ocppj-go/websocket"
	wsclientopt "github.com/e-flux-platform/ocppj-go/websocket/clientopt"
)

// Client represents a client party in the OCPP-J context. It can be used from both the client and server context, i.e:
// - when running a server, each connected party is represented as a client
// - a client can be used to get connected to a running server (e.g. mocking a charging station for testing purposes)
type Client struct {
	id               string
	metadataMux      sync.RWMutex
	metadata         map[string]any
	hooks            ClientHooks
	wsClient         WebsocketClient
	cancel           context.CancelFunc
	outboundCalls    chan outboundCall
	callReplies      chan message.Message
	callTimeout      time.Duration
	processCallsDone chan struct{}
	rateLimiter      ClientRateLimiter
	logger           *slog.Logger
}

// outboundCall represents a request to a send a call to the other party. A context is carried within it; if this
// context is cancelled, the processing of the call will be abandoned. The result of the call is sent on the
// result channel during processing.
type outboundCall struct {
	call    message.Call
	ctx     context.Context
	timeout chan struct{}
	result  chan outboundCallResult
	async   bool
	onDone  func()
}

// outboundCallResult is used to supply the result of an outbound call.
type outboundCallResult struct {
	callReply message.Message // either a message.CallResult or message.CallError
	error     error
}

type CallTimeoutError struct {
	message string
}

func (c CallTimeoutError) Error() string {
	return c.message
}

// CallError is a simple wrapper around message.CallError which implements the error interface
type CallError struct {
	message.CallError
}

func (ce *CallError) Error() string {
	return fmt.Sprintf("call error: %v", ce.CallError)
}

type WebsocketClient interface {
	ID() string
	Read() (websocket.Message, error)
	Write(websocket.Message) error
	Subprotocol() string
	LocalAddr() net.Addr
	RemoteHost() string
	GracefulClose(ctx context.Context, text string) error
	io.Closer
}

// Open creates a connection to an OCPP-J server.
func Open(ctx context.Context, url string, hooks ClientHooks, opts ...ClientOption) (*Client, error) {
	conf := newClientConfig(opts)
	wsSubprotocols := make([]string, 0, len(conf.SupportedProtocols))
	for _, protocol := range conf.SupportedProtocols {
		wsSubprotocols = append(wsSubprotocols, protocol)
	}
	var wsClient WebsocketClient
	if conf.WebsocketClient != nil {
		wsClient = conf.WebsocketClient
	} else {
		var err error
		wsClient, err = websocket.Open(
			ctx,
			url,
			wsclientopt.WithSubprotocols(wsSubprotocols),
			wsclientopt.WithTLSConfig(conf.TLSConfig),
			wsclientopt.WithReadTimeout(conf.WebsocketReadTimeout),
			wsclientopt.WithWriteTimeout(conf.WebsocketWriteTimeout),
			wsclientopt.WithPingInterval(conf.WebsocketPingInterval),
		)
		if err != nil {
			return nil, err
		}
	}
	return newClient(
		uuid.NewString(),
		conf.Metadata,
		hooks,
		wsClient,
		conf.CallTimeout,
		conf.RateLimiter,
		conf.Logger,
	)
}

func newClient(
	id string,
	metadata map[string]any,
	handler ClientHooks,
	wsClient WebsocketClient,
	callTimeout time.Duration,
	rateLimiter ClientRateLimiter,
	logger *slog.Logger,
) (*Client, error) {
	processCtx, cancel := context.WithCancel(context.Background())
	client := &Client{
		id:               id,
		metadata:         metadata,
		wsClient:         wsClient,
		hooks:            handler,
		cancel:           cancel,
		outboundCalls:    make(chan outboundCall),
		callReplies:      make(chan message.Message),
		callTimeout:      callTimeout,
		processCallsDone: make(chan struct{}),
		rateLimiter:      rateLimiter,
		logger:           logger,
	}
	go func() {
		_ = client.processOutboundCallsAndReplies(processCtx)
	}()
	return client, nil
}

// ID returns the unique client identifier
func (c *Client) ID() string {
	return c.id
}

// Read performs a blocking read of messages from the other party. Hooks are invoked for each respective message type.
func (c *Client) Read(ctx context.Context) error {
	for {
		data, err := c.wsClient.Read()
		if err != nil {
			return err
		}

		msg, err := message.FromJSON(data.Data)
		if err != nil {
			if err = c.hooks.OnInvalidMessageRead(ctx, c, data.Data, err); err != nil {
				c.logger.Error("failed to handle invalid message", slog.Any("error", err))
			}
			continue
		}

		if c.hasReachedRateLimit(ctx, msg) {
			c.handleReachedRateLimit(ctx, msg)

			continue
		}

		var isReply bool
		switch m := msg.(type) {
		case message.Call:
			c.logger.Debug("call received", slog.String("client", c.id), slog.Any("payload", m))
			if err = c.hooks.OnCallRead(ctx, c, m); err != nil {
				c.logger.Error("call handling failed", slog.Any("error", err), slog.String("client", c.id), slog.Any("call", m))
			}
		case message.CallResult:
			c.logger.Debug("call result received", slog.String("client", c.id), slog.Any("payload", m))
			if err = c.hooks.OnCallResultRead(ctx, c, m); err != nil {
				c.logger.Error("call result handling failed", slog.Any("error", err), slog.String("client", c.id), slog.Any("callResult", m))
			}
			isReply = true
		case message.CallError:
			c.logger.Debug("call error received", slog.String("client", c.id), slog.Any("payload", m))
			if err = c.hooks.OnCallErrorRead(ctx, c, m); err != nil {
				c.logger.Error("call error handling failed", slog.Any("error", err), slog.String("client", c.id), slog.Any("callError", m))
			}
			isReply = true
		case message.CallResultError:
			c.logger.Debug("call result error received", slog.String("client", c.id), slog.Any("payload", m))
			if err = c.hooks.OnCallResultErrorRead(ctx, c, m); err != nil {
				c.logger.Error("call result error handling failed", slog.Any("error", err), slog.String("client", c.id), slog.Any("callResultError", m))
			}
		case message.Send:
			c.logger.Debug("send received", slog.String("client", c.id), slog.Any("payload", m))
			if err = c.hooks.OnSendRead(ctx, c, m); err != nil {
				c.logger.Error("send handling failed", slog.Any("error", err), slog.String("client", c.id), slog.Any("send", m))
			}
		}

		if isReply {
			// Notify processOutboundCallsAndReplies about the reply
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-c.processCallsDone:
			case c.callReplies <- msg:
			}
		}
	}
}

// WriteCall dispatches a call to the other party. It does not wait for a Call Result or Call Error before returning.
func (c *Client) WriteCall(ctx context.Context, call message.Call) error {
	c.logger.Debug("sending call", slog.String("client", c.id), slog.Any("payload", call))
	_, err := c.sendCall(ctx, call, true)
	return err
}

// SyncWriteCall is a synchronous implementation of WriteCall - it dispatches a Call and then awaits a response. If a
// Call Result is received, this is returned from the function. If a Call Error is received, this is transformed into a
// Go flavour error.
func (c *Client) SyncWriteCall(ctx context.Context, call message.Call) (message.CallResult, error) {
	reply, err := c.sendCall(ctx, call, false)
	if err != nil {
		return message.CallResult{}, err
	}

	switch r := reply.(type) {
	case message.CallResult:
		return r, nil
	case message.CallError:
		return message.CallResult{}, &CallError{CallError: r}
	default:
		return message.CallResult{}, fmt.Errorf("unexpected reply type: %T", reply)
	}
}

// WriteCallResult writes a Call Result type message.
func (c *Client) WriteCallResult(ctx context.Context, callResult message.CallResult) error {
	c.logger.Debug("sending call result", slog.String("client", c.id), slog.Any("payload", callResult))
	if err := c.writeMessage(callResult); err != nil {
		return fmt.Errorf("failed to write call: %w", err)
	}
	if err := c.hooks.OnCallResultWritten(ctx, c, callResult); err != nil {
		return fmt.Errorf("failed to invoke call result written hook: %w", err)
	}
	return nil
}

// WriteCallError writes a Call Error type message.
func (c *Client) WriteCallError(ctx context.Context, callError message.CallError) error {
	c.logger.Debug("sending call error", slog.String("client", c.id), slog.Any("payload", callError))
	if err := c.writeMessage(callError); err != nil {
		return fmt.Errorf("failed to write call error: %w", err)
	}
	if err := c.hooks.OnCallErrorWritten(ctx, c, callError); err != nil {
		return fmt.Errorf("failed to invoke call error written hook: %w", err)
	}
	return nil
}

// WriteCallResultError writes a Call Result Error type message.
func (c *Client) WriteCallResultError(ctx context.Context, callResultError message.CallResultError) error {
	c.logger.Debug("sending call result error", slog.String("client", c.id), slog.Any("payload", callResultError))
	if err := c.writeMessage(callResultError); err != nil {
		return fmt.Errorf("failed to write call result error: %w", err)
	}
	if err := c.hooks.OnCallResultErrorWritten(ctx, c, callResultError); err != nil {
		return fmt.Errorf("failed to invoke call result error written hook: %w", err)
	}
	return nil
}

// WriteSend dispatches a send to the other party.
func (c *Client) WriteSend(ctx context.Context, send message.Send) error {
	c.logger.Debug("sending send", slog.String("client", c.id), slog.Any("payload", send))
	if err := c.writeMessage(send); err != nil {
		return fmt.Errorf("failed to write send: %w", err)
	}
	if err := c.hooks.OnSendWritten(ctx, c, send); err != nil {
		return fmt.Errorf("failed to invoke send written hook: %w", err)
	}
	return nil
}

// Protocol returns the OCPP protocol version
func (c *Client) Protocol() string {
	return c.wsClient.Subprotocol()
}

// Metadata returns a specific metadata value
func (c *Client) Metadata(key string) (any, bool) {
	c.metadataMux.RLock()
	defer c.metadataMux.RUnlock()

	val, found := c.metadata[key]
	return val, found
}

// SetMetadata replaces a metadata key with the supplied value
func (c *Client) SetMetadata(key string, value any) {
	c.metadataMux.Lock()
	defer c.metadataMux.Unlock()

	c.metadata[key] = value
}

// LocalAddr returns the local address of the client
func (c *Client) LocalAddr() net.Addr {
	return c.wsClient.LocalAddr()
}

// RemoteHost returns the host the client is connected to
func (c *Client) RemoteHost() string {
	return c.wsClient.RemoteHost()
}

// WebsocketID returns a unique identifier for the underlying websocket connection
func (c *Client) WebsocketID() string {
	return c.wsClient.ID()
}

// GracefulClose attempts to gracefully close the connection to the other party
func (c *Client) GracefulClose(ctx context.Context, reason string) error {
	c.cancel()
	return c.wsClient.GracefulClose(ctx, reason)
}

// Close performs non-graceful connection termination
func (c *Client) Close() error {
	c.cancel()
	return c.wsClient.Close()
}

// writeMessage writes a raw json message to the underlying websocket
func (c *Client) writeMessage(message any) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return c.wsClient.Write(websocket.TextMessage(payload))
}

// sendCall sends a call to the processOutboundCallsAndReplies function for dispatch. It waits for a reply from the
// processOutboundCallsAndReplies function before returning. NOTE: if async is set to true, then the returned message is
// always nil.
func (c *Client) sendCall(ctx context.Context, call message.Call, async bool) (message.Message, error) {
	result := make(chan outboundCallResult)
	timeout := make(chan struct{})

	timer := time.AfterFunc(c.callTimeout, func() { close(timeout) })

	out := outboundCall{
		call:    call,
		ctx:     ctx,
		timeout: timeout,
		result:  result,
		async:   async,
		onDone:  func() { timer.Stop() },
	}

	// Send to the processOutboundCallsAndReplies function
	select {
	case <-ctx.Done():
		out.onDone()
		return nil, ctx.Err()
	case <-timeout:
		out.onDone()
		return nil, CallTimeoutError{
			message: fmt.Sprintf(
				"write to outbound call channel for message %s timed out after %s",
				call.MessageID,
				c.callTimeout,
			),
		}
	case <-c.processCallsDone:
		out.onDone()
		return nil, fmt.Errorf("client is no longer processing calls")
	case c.outboundCalls <- out:
		// Call is being handled. From this point onwards the invocation of out.onDone() is expected to be handled by
		// the processOutboundCall function - the primary reason being, when async is set to true, this function will
		// return before we've received a reply - and therefore the timeout timer must remain active.
	}

	// Wait for a result back from processOutboundCallsAndReplies
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timeout:
		return nil, CallTimeoutError{
			message: fmt.Sprintf(
				"timed out waiting for a reply for message %s after %s",
				out.call.MessageID,
				c.callTimeout,
			),
		}
	case <-c.processCallsDone:
		return nil, fmt.Errorf("client is no longer processing calls")
	case res := <-result:
		return res.callReply, res.error
	}
}

// processOutboundCallsAndReplies guarantees that only one call is being handled by the other party at any one point in
// time. It sends each call, and waits for either a reply (in the form of a call result or call error), or a timeout.
// This to comply with section 4.1.1. Synchronicity of the 2.0.1 OCPP-J specification which states:
//
//	A Charging Station or CSMS SHALL NOT send a CALL message to the other party unless all the CALL messages it sent
//	before have been responded to or have timed out. This does not mean that the CSMS cannot send a message to another
//	Charging Station, while waiting for a response of a first Charging Station, this rule is per OCPP-J connection. A
//	CALL message has been responded to when a CALLERROR or CALLRESULT message has been received with the message ID of
//	the CALL message.
func (c *Client) processOutboundCallsAndReplies(ctx context.Context) error {
	defer close(c.processCallsDone)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out := <-c.outboundCalls:
			c.processOutboundCall(ctx, out)
		case <-c.callReplies:
			// Out of band call replies are discarded
		}
	}
}

// processOutboundCall processes a single outbound call request.
func (c *Client) processOutboundCall(ctx context.Context, out outboundCall) {
	// The onDone function can be used to take care of any cleanup
	defer out.onDone()

	// Write call via the websocket client
	err := c.writeOutboundCall(ctx, out)
	if err != nil {
		c.notifySender(ctx, out, outboundCallResult{error: err})
		return // If the message write did not appear to succeed, we won't wait for a reply
	}

	// If the sender doesn't require the call result / call error information, we can reply straight away
	if out.async {
		c.notifySender(ctx, out, outboundCallResult{error: nil})
		// Since this is async, from this point onwards we're no longer in the context of the original call
		out.ctx = context.Background()
	}

	// Await a reply (or timeout if no reply appears)
	reply, err := c.awaitOutboundCallReply(ctx, out)

	// Notify sender with the call result / error
	if !out.async {
		c.notifySender(ctx, out, outboundCallResult{callReply: reply, error: err})
	}
}

// writeOutboundCall writes the call message to the underlying websocket
func (c *Client) writeOutboundCall(ctx context.Context, out outboundCall) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-out.ctx.Done():
		return out.ctx.Err()
	case <-out.timeout:
		return CallTimeoutError{
			message: fmt.Sprintf(
				"aborting write to websocket, call with message id %s has timed out after %s",
				out.call.MessageID,
				c.callTimeout,
			),
		}
	default:
		if err := c.writeMessage(out.call); err != nil {
			return fmt.Errorf("failed to write call: %w", err)
		}
		if err := c.hooks.OnCallWritten(out.ctx, c, out.call); err != nil {
			return fmt.Errorf("failed to invoke call written hook: %w", err)
		}
		return nil
	}
}

// awaitOutboundCallReply waits until a reply matching our call is received, a timeout occurs, or a context becomes
// cancelled.
func (c *Client) awaitOutboundCallReply(ctx context.Context, out outboundCall) (message.Message, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-out.ctx.Done():
			return nil, out.ctx.Err()
		case <-out.timeout:
			return nil, CallTimeoutError{
				message: fmt.Sprintf(
					"timed out waiting for call reply for message id %s after %s",
					out.call.MessageID,
					c.callTimeout,
				),
			}
		case reply, ok := <-c.callReplies:
			if !ok {
				return nil, errors.New("reply channel is closed")
			}
			var messageID string
			switch r := reply.(type) {
			case message.CallResult:
				messageID = r.MessageID
			case message.CallError:
				messageID = r.MessageID
			default:
				return nil, fmt.Errorf("unexpected reply type: %T", reply)
			}
			if messageID == out.call.MessageID {
				return reply, nil
			}
		}
	}
}

// notifySender sends an outbound call result back to the sender
func (c *Client) notifySender(ctx context.Context, out outboundCall, result outboundCallResult) {
	select {
	case <-ctx.Done():
	case <-out.ctx.Done():
	case <-out.timeout:
	case out.result <- result:
	}
}

// hasReachedRateLimit returns if the client has reached the rate limit
func (c *Client) hasReachedRateLimit(ctx context.Context, msg message.Message) bool {
	return c.rateLimiter != nil && !c.rateLimiter.Allow(ctx, c, msg)
}

// handleReachedRateLimit handles the case when the client has reached the rate limit in case it's a call message, it
// will send a call error message back to the client otherwise it will do nothing
func (c *Client) handleReachedRateLimit(ctx context.Context, msg message.Message) {
	logger := c.logger.With(slog.String("client", c.id))
	if call, ok := msg.(message.Call); ok {
		err := c.WriteCallError(ctx, message.CallError{
			MessageID:        call.MessageID,
			ErrorCode:        string(message.GenericError),
			ErrorDescription: "rate limit reached",
			ErrorDetails:     []byte("{}"),
		})
		if err != nil {
			logger.Error("failed sending rate limit reached call error")
		}

		logger = logger.With(
			slog.String("message_id", call.MessageID),
			slog.String("action", call.Action),
		)
	}
	logger.Warn("rate limit reached")
}
