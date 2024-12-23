package ocppj

import (
	"context"
	"net/http"
)

type ServerHooks interface {
	OnUpgradeRequested(ctx context.Context, req *http.Request, selectedProtocol string) (*UpgradeRequestResult, error)
	OnClientConnected(context.Context, *Client) error
	OnClientDisconnected(context.Context, *Client) error
}

type UpgradeRequestResult struct {
	ClientID string
	Metadata map[string]any
}
