package ocppj_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/e-flux-platform/ocppj-go/ocppj"
	"github.com/e-flux-platform/ocppj-go/ocppj/clientopt"
	"github.com/e-flux-platform/ocppj-go/ocppj/internal/mocks"
	"github.com/e-flux-platform/ocppj-go/ocppj/message"
	"github.com/e-flux-platform/ocppj-go/websocket"
)

type event struct {
	messageID string
	eventType eventType
}

type eventType int

const (
	eventTypeMessageWritten = iota + 1
	eventTypeCallCompleted
)

func TestClient_WriteCall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("does not write the call until all other active calls have been replied to", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		reads := make(chan websocket.Message)
		wsClient := mocks.NewMockWebsocketClient(ctrl)
		wsClient.EXPECT().Read().AnyTimes().DoAndReturn(func() (websocket.Message, error) {
			return <-reads, nil
		})

		clientHooks := mocks.NewMockClientHooks(ctrl)
		clientHooks.EXPECT().OnCallWritten(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		clientHooks.EXPECT().OnCallResultRead(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		client, err := ocppj.Open(
			ctx,
			"ws://localhost",
			clientHooks,
			clientopt.WithWebsocketClient(wsClient),
			clientopt.WithCallTimeout(time.Minute),
		)
		require.NoError(t, err)

		go func() {
			_ = client.Read(ctx)
		}()

		call1 := message.Call{
			MessageID: "message-1",
			Action:    "GetConfiguration",
			Payload:   json.RawMessage("{}"),
		}
		call2 := message.Call{
			MessageID: "message-2",
			Action:    "GetConfiguration",
			Payload:   json.RawMessage("{}"),
		}

		events := make(chan event)

		wsClient.EXPECT().Write(websocket.TextMessage([]byte(`[2,"message-1","GetConfiguration",{}]`))).DoAndReturn(func(_ any) any {
			events <- event{messageID: call1.MessageID, eventType: eventTypeMessageWritten}
			return nil
		})

		go func() {
			// Write first message
			err := client.WriteCall(ctx, call1)
			assert.NoError(t, err)
			events <- event{messageID: call1.MessageID, eventType: eventTypeCallCompleted}
		}()

		// Message should be written straight away, as we have no pending calls
		assert.Equal(t, event{messageID: call1.MessageID, eventType: eventTypeMessageWritten}, <-events)

		go func() {
			// Write second message
			err := client.WriteCall(ctx, call2)
			assert.NoError(t, err)
			events <- event{messageID: call2.MessageID, eventType: eventTypeCallCompleted}
		}()

		wsClient.EXPECT().Write(websocket.TextMessage([]byte(`[2,"message-2","GetConfiguration",{}]`))).DoAndReturn(func(_ any) any {
			events <- event{messageID: call2.MessageID, eventType: eventTypeMessageWritten}
			return nil
		})

		// Reply to first message
		reads <- websocket.TextMessage([]byte(`[3,"message-1",{}]`))

		// This is the key part - the second message is not written until we've handled the reply for the first message
		assert.Equal(t, event{messageID: call1.MessageID, eventType: eventTypeCallCompleted}, <-events)
		assert.Equal(t, event{messageID: call2.MessageID, eventType: eventTypeMessageWritten}, <-events)

		// Reply to second message
		reads <- websocket.TextMessage([]byte(`[3,"message-2",{}]`))

		assert.Equal(t, event{messageID: call2.MessageID, eventType: eventTypeCallCompleted}, <-events)
	})

	t.Run("safely handles a close when a write has just started", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wsClient := mocks.NewMockWebsocketClient(ctrl)
		wsClient.EXPECT().Write(gomock.Any()).AnyTimes()
		wsClient.EXPECT().Close().AnyTimes()

		clientHooks := mocks.NewMockClientHooks(ctrl)
		clientHooks.EXPECT().OnCallWritten(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		client, err := ocppj.Open(
			ctx,
			"ws://localhost",
			clientHooks,
			clientopt.WithWebsocketClient(wsClient),
			clientopt.WithCallTimeout(time.Minute),
		)
		require.NoError(t, err)

		call1 := message.Call{
			MessageID: "message-1",
			Action:    "GetConfiguration",
			Payload:   json.RawMessage("{}"),
		}

		go func() {
			_ = client.WriteCall(ctx, call1)
		}()
		go func() {
			_ = client.Close()
		}()
	})
}

func TestClient_SyncWriteCall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clientHooks := mocks.NewMockClientHooks(ctrl)
	clientHooks.EXPECT().OnCallWritten(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	clientHooks.EXPECT().OnCallResultRead(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	clientHooks.EXPECT().OnCallErrorRead(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	t.Run("returns a call result when received", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		reads := make(chan websocket.Message)
		wsClient := mocks.NewMockWebsocketClient(ctrl)
		wsClient.EXPECT().Read().AnyTimes().DoAndReturn(func() (websocket.Message, error) {
			return <-reads, nil
		})

		client, err := ocppj.Open(
			ctx,
			"ws://localhost",
			clientHooks,
			clientopt.WithWebsocketClient(wsClient),
			clientopt.WithCallTimeout(time.Minute),
		)
		require.NoError(t, err)

		go func() {
			_ = client.Read(ctx)
		}()

		call := message.Call{
			MessageID: "message",
			Action:    "GetConfiguration",
			Payload:   json.RawMessage("{}"),
		}

		events := make(chan event)

		wsClient.EXPECT().Write(websocket.TextMessage([]byte(`[2,"message","GetConfiguration",{}]`))).DoAndReturn(func(_ any) any {
			events <- event{messageID: call.MessageID, eventType: eventTypeMessageWritten}
			return nil
		})

		results := make(chan message.CallResult)
		go func() {
			res, err := client.SyncWriteCall(ctx, call)
			assert.NoError(t, err)
			results <- res
		}()

		assert.Equal(t, event{messageID: call.MessageID, eventType: eventTypeMessageWritten}, <-events)

		reads <- websocket.TextMessage([]byte(`[3,"message",{}]`))

		actual := <-results
		expected := message.CallResult{
			MessageID: call.MessageID,
			Payload:   json.RawMessage("{}"),
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("returns an error when a call error is received", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		reads := make(chan websocket.Message)
		wsClient := mocks.NewMockWebsocketClient(ctrl)
		wsClient.EXPECT().Read().AnyTimes().DoAndReturn(func() (websocket.Message, error) {
			return <-reads, nil
		})

		client, err := ocppj.Open(
			ctx,
			"ws://localhost",
			clientHooks,
			clientopt.WithWebsocketClient(wsClient),
			clientopt.WithCallTimeout(time.Minute),
		)
		require.NoError(t, err)

		go func() {
			_ = client.Read(ctx)
		}()

		call := message.Call{
			MessageID: "message",
			Action:    "GetConfiguration",
			Payload:   json.RawMessage("{}"),
		}

		events := make(chan event)

		wsClient.EXPECT().Write(websocket.TextMessage([]byte(`[2,"message","GetConfiguration",{}]`))).DoAndReturn(func(_ any) any {
			events <- event{messageID: call.MessageID, eventType: eventTypeMessageWritten}
			return nil
		})

		errors := make(chan error)
		go func() {
			_, err := client.SyncWriteCall(ctx, call)
			errors <- err
		}()

		assert.Equal(t, event{messageID: call.MessageID, eventType: eventTypeMessageWritten}, <-events)

		reads <- websocket.TextMessage([]byte(`[4,"message","GenericError","desc",{}]`))

		actual := <-errors
		expected := &ocppj.CallError{
			CallError: message.CallError{
				MessageID:        call.MessageID,
				ErrorCode:        "GenericError",
				ErrorDescription: "desc",
				ErrorDetails:     json.RawMessage("{}"),
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("it times out cleanly", func(t *testing.T) {
		ctx := context.Background()

		wsClient := mocks.NewMockWebsocketClient(ctrl)
		wsClient.EXPECT().Write(gomock.Any()).Return(nil)

		client, err := ocppj.Open(
			ctx,
			"ws://localhost",
			clientHooks,
			clientopt.WithWebsocketClient(wsClient),
			clientopt.WithCallTimeout(time.Second),
		)
		require.NoError(t, err)

		// We avoid calling the Read function, so no replies will be processed
		_, err = client.SyncWriteCall(ctx, message.Call{
			MessageID: "message",
			Action:    "GetConfiguration",
			Payload:   json.RawMessage("{}"),
		})
		assert.ErrorContains(t, err, "timed out")
	})
}

func TestClient_RateLimit_DoesNotReturnRateLimitCall(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	clientHooks := mocks.NewMockClientHooks(ctrl)
	mockClientRateLimiter := mocks.NewMockClientRateLimiter(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	times := 3

	wg := sync.WaitGroup{}
	wg.Add(times)

	mockClientRateLimiter.EXPECT().Allow(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ any, _ any, _ any) bool {
		wg.Done()
		return true
	}).Times(times)
	clientHooks.EXPECT().OnCallRead(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(times)

	reads := make(chan websocket.Message)
	wsClient := mocks.NewMockWebsocketClient(ctrl)
	wsClient.EXPECT().Read().AnyTimes().DoAndReturn(func() (websocket.Message, error) {
		return <-reads, nil
	})

	client, err := ocppj.Open(
		ctx,
		"ws://localhost",
		clientHooks,
		clientopt.WithWebsocketClient(wsClient),
		clientopt.WithCallTimeout(time.Minute),
		clientopt.WithRateLimiter(mockClientRateLimiter),
	)
	require.NoError(t, err)

	go func() {
		_ = client.Read(ctx)
	}()

	for i := 0; i < times; i++ {
		reads <- websocket.TextMessage([]byte(fmt.Sprintf(`[2,"message-%d","Heartbeat",{}]`, i)))
	}

	wg.Wait()
}

func TestClient_RateLimit_DoesNotReturnRateLimitCallResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clientHooks := mocks.NewMockClientHooks(ctrl)
	mockClientRateLimiter := mocks.NewMockClientRateLimiter(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	times := 3

	wg := sync.WaitGroup{}
	wg.Add(times)

	mockClientRateLimiter.EXPECT().Allow(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ any, _ any, _ any) bool {
		wg.Done()
		return true
	}).Times(times)
	clientHooks.EXPECT().OnCallResultRead(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(times)

	reads := make(chan websocket.Message)
	wsClient := mocks.NewMockWebsocketClient(ctrl)
	wsClient.EXPECT().Read().AnyTimes().DoAndReturn(func() (websocket.Message, error) {
		return <-reads, nil
	})

	client, err := ocppj.Open(
		ctx,
		"ws://localhost",
		clientHooks,
		clientopt.WithWebsocketClient(wsClient),
		clientopt.WithCallTimeout(time.Minute),
		clientopt.WithRateLimiter(mockClientRateLimiter),
	)
	require.NoError(t, err)

	go func() {
		_ = client.Read(ctx)
	}()

	for i := 0; i < times; i++ {
		reads <- websocket.TextMessage([]byte(fmt.Sprintf(`[3,"message-%d",{}]`, i)))
	}

	wg.Wait()
}

func TestClient_RateLimit_ReturnRateLimitCall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clientHooks := mocks.NewMockClientHooks(ctrl)
	mockClientRateLimiter := mocks.NewMockClientRateLimiter(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	times := 3

	wg := sync.WaitGroup{}
	wg.Add(times)

	mockClientRateLimiter.EXPECT().Allow(gomock.Any(), gomock.Any(), gomock.Any()).Return(false).Times(times)
	clientHooks.EXPECT().OnCallErrorWritten(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(times)

	reads := make(chan websocket.Message)
	wsClient := mocks.NewMockWebsocketClient(ctrl)
	wsClient.EXPECT().Read().AnyTimes().DoAndReturn(func() (websocket.Message, error) {
		return <-reads, nil
	})
	wsClient.EXPECT().Write(gomock.Any()).Times(times).DoAndReturn(func(_ any) any {
		wg.Done()
		return nil
	})

	client, err := ocppj.Open(
		ctx,
		"ws://localhost",
		clientHooks,
		clientopt.WithWebsocketClient(wsClient),
		clientopt.WithCallTimeout(time.Minute),
		clientopt.WithRateLimiter(mockClientRateLimiter),
	)
	require.NoError(t, err)

	go func() {
		_ = client.Read(ctx)
	}()

	for i := 0; i < times; i++ {
		reads <- websocket.TextMessage([]byte(fmt.Sprintf(`[2,"message-%d","Heartbeat",{}]`, i)))
	}

	wg.Wait()
}

func TestClient_RateLimit_ReturnRateLimitCallResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clientHooks := mocks.NewMockClientHooks(ctrl)
	mockClientRateLimiter := mocks.NewMockClientRateLimiter(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	times := 3

	wg := sync.WaitGroup{}
	wg.Add(times)

	mockClientRateLimiter.EXPECT().Allow(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ any, _ any, _ any) bool {
		wg.Done()
		return false
	}).Times(times)

	reads := make(chan websocket.Message)
	wsClient := mocks.NewMockWebsocketClient(ctrl)
	wsClient.EXPECT().Read().AnyTimes().DoAndReturn(func() (websocket.Message, error) {
		return <-reads, nil
	})

	client, err := ocppj.Open(
		ctx,
		"ws://localhost",
		clientHooks,
		clientopt.WithWebsocketClient(wsClient),
		clientopt.WithCallTimeout(time.Minute),
		clientopt.WithRateLimiter(mockClientRateLimiter),
	)
	require.NoError(t, err)

	go func() {
		_ = client.Read(ctx)
	}()

	for i := 0; i < times; i++ {
		reads <- websocket.TextMessage([]byte(fmt.Sprintf(`[3,"message-%d",{}]`, i)))
	}

	wg.Wait()
}
