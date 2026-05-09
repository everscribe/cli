package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/everscribe/cli/internal/types"
)

func TestIssueDeviceCode_DecodesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/v1/cli/device-codes", r.URL.Path)
		_ = json.NewEncoder(w).Encode(types.IssueDeviceCodeResponse{
			UserCode:                "ABCD-EFGH",
			DeviceCode:              "long-device-code",
			VerificationURI:         "https://everscribe.io/cli/verify",
			VerificationURIComplete: "https://everscribe.io/cli/verify?user_code=ABCD-EFGH",
			ExpiresIn:               600,
			Interval:                5,
		})
	}))
	t.Cleanup(srv.Close)

	c := New("", WithBaseURL(srv.URL))
	resp, err := c.IssueDeviceCode(context.Background())
	require.NoError(t, err)
	require.Equal(t, "ABCD-EFGH", resp.UserCode)
	require.Equal(t, "long-device-code", resp.DeviceCode)
	require.Equal(t, 600, resp.ExpiresIn)
	require.Equal(t, 5, resp.Interval)
}

func TestExchangeDeviceCode_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req types.ExchangeDeviceCodeRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		require.Equal(t, "long-device-code", req.DeviceCode)
		_ = json.NewEncoder(w).Encode(types.ExchangeDeviceCodeResponse{
			Plaintext: "pat_secret",
			PATID:     "pat-1",
			UserID:    "u_1",
			UserEmail: "alice@example.com",
		})
	}))
	t.Cleanup(srv.Close)

	c := New("", WithBaseURL(srv.URL))
	resp, err := c.ExchangeDeviceCode(context.Background(), "long-device-code")
	require.NoError(t, err)
	require.Equal(t, "pat_secret", resp.Plaintext)
	require.Equal(t, "alice@example.com", resp.UserEmail)
}

func TestExchangeDeviceCode_PendingReturnsTypedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"authorization_pending"}`))
	}))
	t.Cleanup(srv.Close)

	c := New("", WithBaseURL(srv.URL))
	_, err := c.ExchangeDeviceCode(context.Background(), "long-device-code")
	require.Error(t, err)
	require.True(t, IsAuthorizationPending(err), "want IsAuthorizationPending, got %v", err)
}

func TestExchangeDeviceCode_SlowDownAlsoCountsAsPending(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"slow_down"}`))
	}))
	t.Cleanup(srv.Close)

	c := New("", WithBaseURL(srv.URL))
	_, err := c.ExchangeDeviceCode(context.Background(), "x")
	require.Error(t, err)
	require.True(t, IsAuthorizationPending(err), "slow_down should trigger keep-polling")
}

func TestExchangeDeviceCode_ExpiredReturnsTypedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"expired_token"}`))
	}))
	t.Cleanup(srv.Close)

	c := New("", WithBaseURL(srv.URL))
	_, err := c.ExchangeDeviceCode(context.Background(), "x")
	require.True(t, IsExpiredToken(err))
	require.False(t, IsAuthorizationPending(err))
}

func TestExchangeDeviceCode_AccessDeniedReturnsTypedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"access_denied"}`))
	}))
	t.Cleanup(srv.Close)

	c := New("", WithBaseURL(srv.URL))
	_, err := c.ExchangeDeviceCode(context.Background(), "x")
	require.True(t, IsAccessDenied(err))
}

func TestExchangeDeviceCode_500FallsThroughToAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	c := New("", WithBaseURL(srv.URL))
	_, err := c.ExchangeDeviceCode(context.Background(), "x")
	require.Error(t, err)
	require.False(t, IsAuthorizationPending(err))
	require.False(t, IsExpiredToken(err))
}
