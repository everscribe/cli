// Package client is the HTTP client used to talk to the Everscribe API.
//
// The base URL is hardcoded to https://api.everscribe.io. An undocumented
// EVERSCRIBE_API_URL_OVERRIDE environment variable is honored so the test
// suite can point the client at a httptest.Server.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// DefaultBaseURL is the production Everscribe API.
const DefaultBaseURL = "https://api.everscribe.io"

const overrideEnv = "EVERSCRIBE_API_URL_OVERRIDE"

func resolveBaseURL() string {
	if v := os.Getenv(overrideEnv); v != "" {
		return v
	}
	return DefaultBaseURL
}

// Client is the Everscribe API client.
type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// Option customizes Client construction.
type Option func(*Client)

// WithHTTPClient overrides the default *http.Client (mainly for tests).
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.httpClient = h }
}

// WithBaseURL overrides the resolved base URL (mainly for tests).
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = u }
}

// New constructs a Client. Pass an empty token for unauthenticated requests
// (e.g. health checks).
func New(token string, opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    resolveBaseURL(),
		token:      token,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// APIError is returned when the API responds with a non-2xx status.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("api error %d (%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("api error %d: %s", e.StatusCode, e.Message)
}

// IsUnauthorized reports whether err is a 401 from the API.
func IsUnauthorized(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusUnauthorized
}

// IsNotFound reports whether err is a 404 from the API.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// Do performs an authenticated request. body, if non-nil, is JSON-marshalled.
// out, if non-nil, is JSON-decoded from the response body on success.
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return readAPIError(resp)
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// readAPIError tries to interpret the response body as a JSON envelope first,
// then falls back to treating it as plain text. The monorepo currently uses
// http.Error (plain text); the JSON path is forward-compatibility.
func readAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	apiErr := &APIError{StatusCode: resp.StatusCode}

	var env struct {
		Code    string `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
		Error   string `json:"error,omitempty"`
	}
	if len(body) > 0 && json.Unmarshal(body, &env) == nil && (env.Message != "" || env.Error != "") {
		apiErr.Code = env.Code
		if env.Message != "" {
			apiErr.Message = env.Message
		} else {
			apiErr.Message = env.Error
		}
		return apiErr
	}

	apiErr.Message = string(bytes.TrimSpace(body))
	if apiErr.Message == "" {
		apiErr.Message = http.StatusText(resp.StatusCode)
	}
	return apiErr
}
