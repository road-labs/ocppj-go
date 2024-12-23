package ocppj_test

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/e-flux-platform/ocppj-go/ocppj"
	"github.com/e-flux-platform/ocppj-go/ocppj/clientopt"
	"github.com/e-flux-platform/ocppj-go/ocppj/message"
	"github.com/e-flux-platform/ocppj-go/ocppj/mocks"
	"github.com/e-flux-platform/ocppj-go/ocppj/serveropt"
	"github.com/e-flux-platform/ocppj-go/websocket"

	"golang.org/x/time/rate"
)

const (
	serverPort      = 9822
	maxNumOfClients = 5000
)

var defaultRateLimiter = rate.NewLimiter(rate.Every(time.Second), 1000000)

// go test -benchmem -benchtime 5000x -run=^$ -bench ^BenchmarkServer_OpenClients$ github.com/e-flux-platform/ocppj-go/ocppj
func BenchmarkServer_OpenClients(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(b.N)
	fmt.Println("")
	slog.Info("starting BenchmarkServer_OpenClients benchmark with", slog.Int("clients", b.N))

	mockServerHooks, mockClientHooks, upgrader := getMocks(b)
	mockServerHooks.EXPECT().OnClientConnected(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, client *ocppj.Client) error {
		wg.Done()
		return nil
	}).AnyTimes()

	server := ocppj.NewServer(
		mockServerHooks,
		mockClientHooks,
		serveropt.WithUpgrader(upgrader),
		serveropt.WithSupportedProtocols([]string{"ocpp2.0.1"}),
		serveropt.WithPort(serverPort),
		serveropt.WithClientRateLimiter(nil),
	)

	go func() {
		err := server.Start(ctx)
		if err != nil {
			panic(err)
		}
	}()

	for i := 0; i < b.N; i++ {
		ocppIdentity := fmt.Sprintf("client-%d", i)
		mockServerHooks.EXPECT().OnUpgradeRequested(gomock.Any(), gomock.Any(), gomock.Any()).Return(&ocppj.UpgradeRequestResult{
			ClientID: ocppIdentity,
			Metadata: map[string]any{},
		}, nil).Times(1)

		client, err := ocppj.Open(
			ctx,
			fmt.Sprintf("ws://localhost:%d/%s", serverPort, ocppIdentity),
			mockClientHooks,
			clientopt.WithSupportedProtocols([]string{"ocpp2.0.1"}),
		)
		if err != nil {
			b.Error(err)
		}

		go func() {
			_ = client.Read(context.Background())
		}()
	}

	slog.Info("waiting for clients to connect")
	wg.Wait()
	slog.Info("all clients connected")
}

// go test -benchmem -benchtime 5000x -run=^$ -bench ^BenchmarkServer_OpenClients_WithRateLimiter$ github.com/e-flux-platform/ocppj-go/ocppj
func BenchmarkServer_OpenClients_WithRateLimiter(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(b.N)
	fmt.Println("")
	slog.Info("starting BenchmarkServer_OpenClients_WithRateLimiter benchmark with", slog.Int("clients", b.N))

	rateLimiter := &fakeRateLimiter{
		defaultRateLimiter: defaultRateLimiter,
		clients:            make(map[string]*rate.Limiter),
	}

	mockServerHooks, mockClientHooks, upgrader := getMocks(b)
	mockServerHooks.EXPECT().OnClientConnected(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, client *ocppj.Client) error {
		rateLimiter.OnClientConnected(client)
		wg.Done()
		return nil
	}).AnyTimes()

	server := ocppj.NewServer(
		mockServerHooks,
		mockClientHooks,
		serveropt.WithUpgrader(upgrader),
		serveropt.WithSupportedProtocols([]string{"ocpp2.0.1"}),
		serveropt.WithPort(serverPort),
		serveropt.WithClientRateLimiter(rateLimiter),
	)

	go func() {
		err := server.Start(ctx)
		if err != nil {
			panic(err)
		}
	}()

	for i := 0; i < b.N; i++ {
		ocppIdentity := fmt.Sprintf("client-%d", i)
		mockServerHooks.EXPECT().OnUpgradeRequested(gomock.Any(), gomock.Any(), gomock.Any()).Return(&ocppj.UpgradeRequestResult{
			ClientID: ocppIdentity,
			Metadata: map[string]any{},
		}, nil).Times(1)

		client, err := ocppj.Open(
			ctx,
			fmt.Sprintf("ws://localhost:%d/%s", serverPort, ocppIdentity),
			mockClientHooks,
			clientopt.WithSupportedProtocols([]string{"ocpp2.0.1"}),
		)
		if err != nil {
			b.Error(err)
		}

		go func() {
			_ = client.Read(context.Background())
		}()
	}

	slog.Info("waiting for clients to connect")
	wg.Wait()
	slog.Info("all clients connected")
}

// go test -benchmem -benchtime 10000x -run=^$ -bench ^BenchmarkServer_SendMessages$ github.com/e-flux-platform/ocppj-go/ocppj
func BenchmarkServer_SendMessages(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(b.N)
	fmt.Println("")
	slog.Info("starting BenchmarkServer_SendMessages benchmark with", slog.Int("clients", maxNumOfClients), slog.Int("messages", b.N))

	mockServerHooks, mockClientHooks, upgrader := getMocks(b)
	mockServerHooks.EXPECT().OnClientConnected(gomock.Any(), gomock.Any()).AnyTimes()
	mockClientHooks.EXPECT().OnCallRead(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, client *ocppj.Client, call message.Call) error {
		_ = client.WriteCallResult(ctx, message.CallResult{
			MessageID: call.MessageID,
			Payload:   []byte(`{"currentTime": "2020-01-01T00:00:00Z"}`),
		})
		wg.Done()
		return nil
	}).AnyTimes()

	server := ocppj.NewServer(
		mockServerHooks,
		mockClientHooks,
		serveropt.WithUpgrader(upgrader),
		serveropt.WithSupportedProtocols([]string{"ocpp2.0.1"}),
		serveropt.WithPort(serverPort),
		serveropt.WithClientRateLimiter(nil),
	)

	go func() {
		err := server.Start(ctx)
		if err != nil {
			panic(err)
		}
	}()

	clients := make([]*ocppj.Client, 0)

	for i := 0; i < maxNumOfClients; i++ {
		ocppIdentity := fmt.Sprintf("client-%d", i)

		mockServerHooks.EXPECT().OnUpgradeRequested(gomock.Any(), gomock.Any(), gomock.Any()).Return(&ocppj.UpgradeRequestResult{
			ClientID: ocppIdentity,
			Metadata: map[string]any{},
		}, nil).Times(1)

		client, err := ocppj.Open(
			ctx,
			fmt.Sprintf("ws://localhost:%d/%s", serverPort, ocppIdentity),
			mockClientHooks,
			clientopt.WithSupportedProtocols([]string{"ocpp2.0.1"}),
		)
		if err != nil {
			b.Error(err)
		}

		go func() {
			_ = client.Read(context.Background())
		}()

		clients = append(clients, client)
	}

	r := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))

	for i := 0; i < b.N; i++ {
		client := clients[r.IntN(maxNumOfClients-1)]

		err := client.WriteCall(context.Background(), message.Call{
			MessageID: fmt.Sprintf("message-%d", i),
			Action:    "Heartbeat",
			Payload:   nil,
		})
		if err != nil {
			b.Error(err)
		}
	}

	slog.Info("waiting for clients to receive messages")
	wg.Wait()
	slog.Info("all clients received messages")
}

// go test -benchmem -benchtime 10000x -run=^$ -bench ^BenchmarkServer_SendMessages_WithRateLimiter$ github.com/e-flux-platform/ocppj-go/ocppj
func BenchmarkServer_SendMessages_WithRateLimiter(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(b.N)
	fmt.Println("")
	slog.Info("starting BenchmarkServer_SendMessages_WithRateLimiter benchmark with", slog.Int("clients", maxNumOfClients), slog.Int("messages", b.N))

	rateLimiter := &fakeRateLimiter{
		defaultRateLimiter: defaultRateLimiter,
		clients:            make(map[string]*rate.Limiter),
	}

	mockServerHooks, mockClientHooks, upgrader := getMocks(b)
	mockServerHooks.EXPECT().OnClientConnected(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, client *ocppj.Client) error {
		rateLimiter.OnClientConnected(client)
		return nil
	}).AnyTimes()

	mockClientHooks.EXPECT().OnCallRead(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, client *ocppj.Client, call message.Call) error {
		_ = client.WriteCallResult(ctx, message.CallResult{
			MessageID: call.MessageID,
			Payload:   []byte(`{"currentTime": "2020-01-01T00:00:00Z"}`),
		})
		wg.Done()
		return nil
	}).AnyTimes()

	server := ocppj.NewServer(
		mockServerHooks,
		mockClientHooks,
		serveropt.WithUpgrader(upgrader),
		serveropt.WithSupportedProtocols([]string{"ocpp2.0.1"}),
		serveropt.WithPort(serverPort),
		serveropt.WithClientRateLimiter(rateLimiter),
	)

	go func() {
		err := server.Start(ctx)
		if err != nil {
			panic(err)
		}
	}()

	clients := []*ocppj.Client{}

	for i := 0; i < maxNumOfClients; i++ {
		ocppIdentity := fmt.Sprintf("client-%d", i)

		mockServerHooks.EXPECT().OnUpgradeRequested(gomock.Any(), gomock.Any(), gomock.Any()).Return(&ocppj.UpgradeRequestResult{
			ClientID: ocppIdentity,
			Metadata: map[string]any{},
		}, nil).Times(1)

		client, err := ocppj.Open(
			ctx,
			fmt.Sprintf("ws://localhost:%d/%s", serverPort, ocppIdentity),
			mockClientHooks,
			clientopt.WithSupportedProtocols([]string{"ocpp2.0.1"}),
		)
		if err != nil {
			b.Error(err)
		}

		go func() {
			_ = client.Read(context.Background())
		}()

		clients = append(clients, client)
	}

	r := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))

	for i := 0; i < b.N; i++ {
		client := clients[r.IntN(maxNumOfClients-1)]

		err := client.WriteCall(context.Background(), message.Call{
			MessageID: fmt.Sprintf("message-%d", i),
			Action:    "Heartbeat",
			Payload:   nil,
		})
		if err != nil {
			b.Error(err)
		}
	}

	slog.Info("waiting for clients to receive messages")
	wg.Wait()
	slog.Info("all clients received messages")
}

func getMocks(t gomock.TestReporter) (*mocks.MockServerHooks, *mocks.MockClientHooks, *websocket.Upgrader) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServerHooks := mocks.NewMockServerHooks(ctrl)

	mockServerHooks.EXPECT().OnClientDisconnected(gomock.Any(), gomock.Any()).AnyTimes()

	mockClientHooks := mocks.NewMockClientHooks(ctrl)

	mockClientHooks.EXPECT().OnCallResultRead(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockClientHooks.EXPECT().OnCallErrorRead(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockClientHooks.EXPECT().OnCallWritten(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockClientHooks.EXPECT().OnCallResultWritten(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockClientHooks.EXPECT().OnCallErrorWritten(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	upgrader := websocket.NewUpgrader()

	return mockServerHooks, mockClientHooks, upgrader
}

type fakeRateLimiter struct {
	mux sync.RWMutex

	defaultRateLimiter *rate.Limiter
	clients            map[string]*rate.Limiter
}

func (crl *fakeRateLimiter) Allow(ctx context.Context, client *ocppj.Client, msg message.Message) bool {
	crl.mux.Lock()
	defer crl.mux.Unlock()

	limiter, found := crl.clients[client.ID()]
	if !found {
		return true
	}

	return limiter.Allow()
}

func (crl *fakeRateLimiter) OnClientConnected(client *ocppj.Client) {
	crl.mux.Lock()
	defer crl.mux.Unlock()

	if _, found := crl.clients[client.ID()]; !found {
		crl.clients[client.ID()] = rate.NewLimiter(crl.defaultRateLimiter.Limit(), crl.defaultRateLimiter.Burst())
	}
}
