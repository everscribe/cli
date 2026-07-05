package client

import (
	"context"
	"net/http"

	"github.com/everscribe/cli/internal/types"
)

func (c *Client) ListAPIKeys(ctx context.Context, projectID string) ([]types.APIKey, error) {
	var out types.ListAPIKeysResponse
	if err := c.Do(ctx, http.MethodGet, "/v1/projects/"+projectID+"/keys", nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Keys, nil
}

// CreateAPIKey returns the plaintext key alongside the metadata -
// the plaintext is only surfaced this once.
func (c *Client) CreateAPIKey(ctx context.Context, projectID, name string) (*types.CreateAPIKeyResponse, error) {
	var out types.CreateAPIKeyResponse
	body := types.CreateAPIKeyRequest{Name: name}
	if err := c.Do(ctx, http.MethodPost, "/v1/projects/"+projectID+"/keys", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) RevokeAPIKey(ctx context.Context, projectID, keyID string) error {
	return c.Do(ctx, http.MethodDelete, "/v1/projects/"+projectID+"/keys/"+keyID, nil, nil, nil)
}
