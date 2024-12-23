package ocppj

import (
	"context"

	"github.com/e-flux-platform/ocppj-go/ocppj/message"
)

type ClientHooks interface {
	// OnCallRead is invoked when a client reads a message from the other party. A reply will typically need to be sent;
	// this can be achieved by calling WriteCall() on the supplied client.
	OnCallRead(context.Context, *Client, message.Call) error
	// OnCallResultRead is invoked when a client reads a call reply from the other party.
	OnCallResultRead(context.Context, *Client, message.CallResult) error
	// OnCallErrorRead is invoked when a client reads a call error from the other party.
	OnCallErrorRead(context.Context, *Client, message.CallError) error
	// OnInvalidMessageRead is invoked when the client receives a message that cannot be parsed into a valid OCPP-J
	// message. Common reasons for this include invalid JSON, and the payload including invalid UTF-8. Note that the
	// behaviour within the read loop is to simply skip past such messages; if there is a desire to disconnect the
	// client instead, the hook implementation should call Close() on the supplied Client instance.
	OnInvalidMessageRead(context.Context, *Client, []byte, error) error
	// OnCallWritten is invoked when the client writes a call. This can be useful for logging, etc, particularly when
	// the WriteCall() function can end up being called from multiple contexts.
	OnCallWritten(context.Context, *Client, message.Call) error
	// OnCallResultWritten is invoked when the client writes a call result. This can be useful for logging, etc,
	// particularly when the WriteCallResult() function can end up being called from multiple contexts.
	OnCallResultWritten(context.Context, *Client, message.CallResult) error
	// OnCallErrorWritten is invoked when the client writes a call error. This can be useful for logging, etc,
	// particularly when the WriteCallError() function can end up being called from multiple contexts.
	OnCallErrorWritten(context.Context, *Client, message.CallError) error
}
