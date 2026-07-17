# ocppj-go

Go implementation of the OCPP-J "JSON over WebSockets" specification.

## Features

This package handles a base level OCPP-J implementation, including:

- Client and server communication over WebSockets, using either plaintext or TLS
- Serialization and parsing of the OCPP-J message envelopes, with basic validation
- Synchronicity enforcement, so that only one outbound call can be in flight at a time in either direction

The code looks to be open to forward extension and not pinned to any specific OCPP versions.

## Usage

This package covers both client and server implementations of the OCPP-J protocol. The client implementation can be used
to establish connections to an OCPP server, which can be useful for implementing charging station simulators, etc. The
server implementation provides the plumbing to build out an OCPP backend.

Callers wire in client and server behaviour through hooks. `ocppj.ClientHooks` covers read and write of each message
type per-client, and `ocppj.ServerHooks` covers the upgrade and connection lifecycle.

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

	client, err := ocppj.Open(
		ctx,
		"ws://localhost:2600/ChargeStationFoo",
		&clientHooks{},
		clientopt.WithSupportedProtocols([]string{"ocpp1.6"}),
		clientopt.WithWebsocketPingInterval(time.Minute),
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

// The remaining ocppj.ClientHooks methods are omitted for brevity. See the interface for the full set.
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

	server, err := ocppj.NewServer(
		&serverHooks{},
		&clientHooks{},
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

// serverHooks handle upgrade requests and client connections/disconnections.
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

// clientHooks implements ocppj.ClientHooks, receiving messages read from each connected client.
type clientHooks struct{}

func (c *clientHooks) OnCallRead(ctx context.Context, client *ocppj.Client, call message.Call) error {
	// Reply via client.WriteCallResult or client.WriteCallError.
	return nil
}

// The remaining ocppj.ClientHooks methods are omitted for brevity. See the interface for the full set.
```

## Testing

```sh
go test ./...
```

Mocks are generated with [`mockgen`](https://github.com/uber-go/mock). See `Taskfile.yaml` for the generation commands.

## License

MIT. See [LICENSE](LICENSE).
