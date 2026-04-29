package ocppj

import (
	"context"

	"github.com/e-flux-platform/ocppj-go/ocppj/message"
)

type ClientHooks interface {
	// OnCallRead is invoked when a client reads a "call" message from the other party. A reply will typically need to be
	// sent; this can be achieved by calling WriteCallResult() or WriteCallError() on the supplied client.
	OnCallRead(context.Context, *Client, message.Call) error
	// OnCallResultRead is invoked when a client reads a "call result" message from the other party. In OCPP >= 2.1 this
	// can be replied to with a call result error, if necessary.
	OnCallResultRead(context.Context, *Client, message.CallResult) error
	// OnCallErrorRead is invoked when a client reads a "call error" message from the other party.
	OnCallErrorRead(context.Context, *Client, message.CallError) error
	// OnCallResultErrorRead is invoked when a client reads a "call result error" message from the other party.
	OnCallResultErrorRead(context.Context, *Client, message.CallResultError) error
	// OnSendRead is invoked when a client reads a "send" message from the other party. No reply is expected.
	OnSendRead(context.Context, *Client, message.Send) error
	// OnInvalidMessageRead is invoked when the client receives a message that cannot be parsed into a valid OCPP-J
	// message. Common reasons for this include invalid JSON, and the payload including invalid UTF-8. Note that the
	// behaviour within the read loop is to simply skip past such messages; if there is a desire to do anything else,
	// this should be handled within the hook.
	OnInvalidMessageRead(context.Context, *Client, []byte, error) error
	// OnCallWritten is invoked when the client writes a call. This can be useful for logging, etc, particularly when
	// the WriteCall() function can end up being called from multiple contexts.
	OnCallWritten(context.Context, *Client, message.Call) error
	// OnCallResultWritten is invoked when the client writes a call result.
	OnCallResultWritten(context.Context, *Client, message.CallResult) error
	// OnCallErrorWritten is invoked when the client writes a call error.
	OnCallErrorWritten(context.Context, *Client, message.CallError) error
	// OnCallResultErrorWritten is invoked when the client writes a call result error.
	OnCallResultErrorWritten(context.Context, *Client, message.CallResultError) error
	// OnSendWritten is invoked when the client writes a send.
	OnSendWritten(context.Context, *Client, message.Send) error
}
