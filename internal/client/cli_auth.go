package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/everscribe/cli/internal/types"
)

// IssueDeviceCode requests a new (user_code, device_code) pair from
// the API. Unauthenticated — the device_code is the credential for
// the subsequent exchange.
func (c *Client) IssueDeviceCode(ctx context.Context) (*types.IssueDeviceCodeResponse, error) {
	var out types.IssueDeviceCodeResponse
	if err := c.Do(ctx, http.MethodPost, "/v1/cli/device-codes", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeviceExchangeError is the typed RFC 8628 error returned by
// ExchangeDeviceCode when the server replies 400 with an error code.
// The login flow checks Code to decide whether to keep polling.
type DeviceExchangeError struct {
	Code string
}

func (e *DeviceExchangeError) Error() string {
	return "device exchange: " + e.Code
}

// IsAuthorizationPending reports whether err is the keep-polling
// signal from the exchange endpoint.
func IsAuthorizationPending(err error) bool {
	var dErr *DeviceExchangeError
	if !errors.As(err, &dErr) {
		return false
	}
	return dErr.Code == types.DeviceErrorAuthorizationPending || dErr.Code == types.DeviceErrorSlowDown
}

// IsExpiredToken reports whether err means the device code is no
// longer usable (expired, revoked, or already consumed).
func IsExpiredToken(err error) bool {
	var dErr *DeviceExchangeError
	return errors.As(err, &dErr) && dErr.Code == types.DeviceErrorExpiredToken
}

// IsAccessDenied reports whether err means the user explicitly
// denied the device.
func IsAccessDenied(err error) bool {
	var dErr *DeviceExchangeError
	return errors.As(err, &dErr) && dErr.Code == types.DeviceErrorAccessDenied
}

// ExchangeDeviceCode polls the exchange endpoint with the device_code.
// On success returns the minted PAT data. On a 400 with an RFC 8628
// error code, returns *DeviceExchangeError. Other errors propagate
// as the underlying APIError or transport error.
func (c *Client) ExchangeDeviceCode(ctx context.Context, deviceCode string) (*types.ExchangeDeviceCodeResponse, error) {
	body, err := json.Marshal(types.ExchangeDeviceCodeRequest{DeviceCode: deviceCode})
	if err != nil {
		return nil, fmt.Errorf("marshal exchange body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/cli/device-tokens", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var out types.ExchangeDeviceCodeResponse
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		return &out, nil
	}

	if resp.StatusCode == http.StatusBadRequest {
		// RFC 8628 error envelope: {"error": "authorization_pending"|...}
		raw, _ := io.ReadAll(resp.Body)
		var env struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(raw, &env) == nil && env.Error != "" {
			return nil, &DeviceExchangeError{Code: env.Error}
		}
		// Fall through if the body wasn't an RFC envelope.
		return nil, &APIError{StatusCode: resp.StatusCode, Message: string(bytes.TrimSpace(raw))}
	}

	// Anything else is a real transport / server error.
	raw, _ := io.ReadAll(resp.Body)
	return nil, &APIError{StatusCode: resp.StatusCode, Message: string(bytes.TrimSpace(raw))}
}
