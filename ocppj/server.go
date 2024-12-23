package ocppj

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"

	"github.com/e-flux-platform/ocppj-go/websocket"
)

const secWebsocketProtocolHeader = "Sec-WebSocket-Protocol"

// Server represents a server party in the OCPP-J context. It runs starts a HTTP server and accepts upgrade requests
// on a configured path.
type Server struct {
	conf             *ServerConfig
	hooks            ServerHooks
	clientHooks      ClientHooks
	wsUpgrader       Upgrader
	shutdownStrategy ShutdownStrategy
	wg               sync.WaitGroup
	mux              sync.Mutex
	clients          map[string]*Client
}

type Upgrader interface {
	Upgrade(w http.ResponseWriter, req *http.Request, responseHeader http.Header) (*websocket.Client, error)
}

// NewServer creates a new server instance
func NewServer(hooks ServerHooks, clientHooks ClientHooks, opts ...ServerOption) (*Server, error) {
	conf := newServerConfig(opts)

	return &Server{
		conf:             conf,
		hooks:            hooks,
		clientHooks:      clientHooks,
		wsUpgrader:       conf.Upgrader,
		shutdownStrategy: conf.ShutdownStrategy,
		clients:          make(map[string]*Client),
	}, nil
}

// Start runs the server. This is a blocking operation, and will not return until the server is interrupted in some way
// (e.g. the supplied context is cancelled)
func (s *Server) Start(ctx context.Context) error {
	r := http.NewServeMux()
	r.HandleFunc(s.conf.HealthPath, s.handleHealth)
	r.HandleFunc(s.conf.UpgradePath, s.handleUpgrade)

	server := &http.Server{
		Addr:              s.conf.ListenAddr,
		TLSConfig:         s.conf.TLSConfig,
		ReadHeaderTimeout: s.conf.ReadHeaderTimeout,
		Handler:           r,
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		slog.Info("ocpp-j server listening", slog.Any("address", s.conf.ListenAddr), slog.Bool("tls", s.conf.TLSConfig != nil))
		var err error
		if s.conf.TLSConfig != nil {
			err = server.ListenAndServeTLS("", "")
		} else {
			err = server.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})
	eg.Go(func() error {
		<-ctx.Done()
		sCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		return server.Shutdown(sCtx) //nolint:contextcheck
	})
	err := eg.Wait()
	s.drainClients() // The server should have stopped listening, so we can start to drain the clients
	return err
}

func (s *Server) handleUpgrade(w http.ResponseWriter, req *http.Request) {
	// Select which protocol we're going to use
	protocol, err := s.selectProtocol(req.Header.Get(secWebsocketProtocolHeader))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Process the upgrade request
	res, err := s.hooks.OnUpgradeRequested(req.Context(), req, protocol)
	if err != nil {
		slog.Error("upgrade request failed", slog.Any("error", err))
		upgradeError := &ServerUpgradeError{}
		if errors.As(err, &upgradeError) {
			http.Error(w, upgradeError.message, upgradeError.httpStatus)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Perform websocket upgrade
	respHeaders := make(http.Header)
	respHeaders.Set(secWebsocketProtocolHeader, string(protocol))
	wsClient, err := s.wsUpgrader.Upgrade(w, req, respHeaders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Spawn a goroutine to handle reading from the client connection
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		ctx := context.Background()

		client, err := newClient(
			res.ClientID,
			res.Metadata,
			s.clientHooks,
			wsClient,
			s.conf.ClientCallTimeout,
			s.conf.ClientRateLimiter,
		)
		if err != nil {
			_ = wsClient.Close()
			slog.Error("failed to create client", slog.Any("error", err))
			return
		}
		defer client.Close()

		err = s.processClient(ctx, client)
		slog.Info(
			"client disconnected",
			slog.String("id", client.ID()),
			slog.String("websocketId", client.WebsocketID()),
			slog.Any("error", err),
		)
	}()
}

// processClient registers the client and then consumes messages from it
func (s *Server) processClient(ctx context.Context, client *Client) error {
	s.mux.Lock()
	s.clients[client.WebsocketID()] = client
	s.mux.Unlock()

	defer func() {
		s.mux.Lock()
		delete(s.clients, client.WebsocketID())
		s.mux.Unlock()
	}()

	if err := s.hooks.OnClientConnected(ctx, client); err != nil {
		return err
	}
	defer func() {
		err := s.hooks.OnClientDisconnected(ctx, client)
		if err != nil {
			slog.Error("client disconnect hook failed", slog.Any("error", err))
		}
	}()

	return client.Read(ctx)
}

// selectProtocol selects the appropriate protocol to use from those supported by the client
func (s *Server) selectProtocol(requestedProtocols string) (string, error) {
	// Parse requested versions string
	parts := strings.Split(requestedProtocols, ",")
	versions := make(map[string]struct{})
	for _, part := range parts {
		version := strings.TrimSpace(part)
		versions[version] = struct{}{}
	}

	for _, v := range s.conf.SupportedProtocols {
		if _, found := versions[v]; found {
			return v, nil
		}
	}

	return "", fmt.Errorf("unsupported protocol version(s): %s", requestedProtocols)
}

func (s *Server) drainClients() {
	s.mux.Lock()
	clients := maps.Values(s.clients)
	s.mux.Unlock()

	slog.Info("draining clients", slog.Int("total", len(clients)))

	// Disconnect all clients
	ctx := context.Background()
	if err := s.shutdownStrategy(ctx, clients); err != nil {
		slog.Error("shutdown failed", slog.Any("error", err))
		return
	}

	// Wait for all client connection goroutines to complete
	s.wg.Wait()
}

func (s *Server) handleHealth(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

type ServerUpgradeError struct {
	httpStatus int
	message    string
}

func (s *ServerUpgradeError) Error() string {
	return s.message
}
