# ocppj-go

Go implementation of the OCPP-J "JSON over WebSockets" specification. This package covers the base level implementation,
including:

- client to client communication over WebSockets, using either plaintext or TLS
- serialization/deserialization of message payloads, including general validation of the contents
- synchronicity enforcement, which guarantees only one outbound message is active at any point in time (both on the
sending and receiving ends)

The implementation looks to be open to forward extension, and not pinned to any specific OCPP versions.

## Overview

This package covers both client and server implementations of the OCPP-J protocol. From a usage perspective, the client
implementation can be used for charging station simulation type usecases, and the server implementation can be used
as the basic rails to build an OCPP backend.

There are certain "non-goals" for this project, namely:

- there is no intention for this to be a framework for OCPP/CSMS implementations. There is a general preference towards
simplicity and flexibility, rather than offering a fully fledged OCPP stack
- the implementation does not look to handle OCPP version specific payloads - these are passed on to the hook
implementations to deal with decoding, handling, etc. There is a tendency to find stations that are non-compliant to
OCPP in the wild, and whilst the temptation is to simply reject payloads from such stations, sometimes there is a need
to be pragmatic and support them.

## Examples

See below for basic client and server examples.

### Client

```go
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/e-flux-platform/ocppj-go/ocppj"
	"github.com/e-flux-platform/ocppj-go/ocppj/clientopt"
	"github.com/e-flux-platform/ocppj-go/ocppj/message"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	hooks := &clientHooks{}

	client, err := ocppj.Open(
		ctx,
		"ws://localhost:2600/ChargeStationFoo",
		hooks,
		clientopt.WithSupportedProtocols([]string{"ocpp1.6"}),
		clientopt.WithWebsocketPingInternal(time.Minute),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	go client.Read(ctx)

	<-ctx.Done()
}

type clientHooks struct{}

func (c *clientHooks) OnCallRead(ctx context.Context, client *ocppj.Client, call message.Call) error {
	// TODO: handle Call read
	return nil
}

func (c *clientHooks) OnCallResultRead(ctx context.Context, client *ocppj.Client, callResult message.CallResult) error {
	// TODO: handle Call Result read
	return nil
}

func (c *clientHooks) OnCallErrorRead(ctx context.Context, client *ocppj.Client, callError message.CallError) error {
	// TODO: handle Call Error read
	return nil
}

func (c *clientHooks) OnInvalidMessageRead(ctx context.Context, client *ocppj.Client, bytes []byte, err error) error {
	// TODO: handle invalid message read
	return nil
}

func (c *clientHooks) OnCallWritten(ctx context.Context, client *ocppj.Client, call message.Call) error {
	// TODO: handle Call written
	return nil
}

func (c *clientHooks) OnCallResultWritten(ctx context.Context, client *ocppj.Client, result message.CallResult) error {
	// TODO: handle Call Result written
	return nil
}

func (c *clientHooks) OnCallErrorWritten(ctx context.Context, client *ocppj.Client, callError message.CallError) error {
	// TODO: handle Call Error written
	return nil
}
```

### Server

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/google/uuid"

	"github.com/e-flux-platform/ocppj-go/ocppj"
	"github.com/e-flux-platform/ocppj-go/ocppj/message"
	"github.com/e-flux-platform/ocppj-go/ocppj/serveropt"
	"github.com/e-flux-platform/ocppj-go/websocket"
	"github.com/e-flux-platform/ocppj-go/websocket/upgraderopt"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	upgrader := websocket.NewUpgrader(
		upgraderopt.WithOriginCheck(func(_ *http.Request) bool {
			return true
		}),
	)

	sHooks := &serverHooks{}
	cHooks := &clientHooks{}

	server, err := ocppj.NewServer(
		sHooks,
		cHooks,
		serveropt.WithPort(2600),
		serveropt.WithSupportedProtocols([]string{"ocpp1.6", "ocpp2.0.1"}),
		serveropt.WithUpgradePath("/{ocppIdentity}"),
		serveropt.WithUpgrader(upgrader),
	)
	if err != nil {
		log.Fatal(err)
	}

	_ = server.Start(ctx)
}

type serverHooks struct{}

func (s *serverHooks) OnUpgradeRequested(ctx context.Context, req *http.Request, selectedProtocol string) (*ocppj.UpgradeRequestResult, error) {
	return &ocppj.UpgradeRequestResult{
		ClientID: uuid.NewString(),
	}, nil
}

func (s *serverHooks) OnClientConnected(ctx context.Context, client *ocppj.Client) error {
	// TODO: handle client connected
	return nil
}

func (s *serverHooks) OnClientDisconnected(ctx context.Context, client *ocppj.Client) error {
	// TODO: handle client disconnected
	return nil
}

type clientHooks struct{}

func (c *clientHooks) OnCallRead(ctx context.Context, client *ocppj.Client, call message.Call) error {
	// TODO: handle Call read
	return nil
}

func (c *clientHooks) OnCallResultRead(ctx context.Context, client *ocppj.Client, callResult message.CallResult) error {
	// TODO: handle Call Result read
	return nil
}

func (c *clientHooks) OnCallErrorRead(ctx context.Context, client *ocppj.Client, callError message.CallError) error {
	// TODO: handle Call Error read
	return nil
}

func (c *clientHooks) OnInvalidMessageRead(ctx context.Context, client *ocppj.Client, bytes []byte, err error) error {
	// TODO: handle invalid message read
	return nil
}

func (c *clientHooks) OnCallWritten(ctx context.Context, client *ocppj.Client, call message.Call) error {
	// TODO: handle Call written
	return nil
}

func (c *clientHooks) OnCallResultWritten(ctx context.Context, client *ocppj.Client, result message.CallResult) error {
	// TODO: handle Call Result written
	return nil
}

func (c *clientHooks) OnCallErrorWritten(ctx context.Context, client *ocppj.Client, callError message.CallError) error {
	// TODO: handle Call Error written
	return nil
}
```