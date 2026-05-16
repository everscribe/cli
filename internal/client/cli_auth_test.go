package client

import (
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
	resp, err := c.IssueDeviceCode(t.Context())
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
	resp, err := c.ExchangeDeviceCode(t.Context(), "long-device-code")
	require.NoError(t, err)
	require.Equal(t, "pat_secret", resp.Plaintext)
	require.Equal(t, "alice@example.com", resp.UserEmail)
}

// TestExchangeDeviceCode_OAuthErrorMapping verifies the 400-with-JSON-error
// branch: each OAuth error code the IETF device-flow spec defines must
// be classified by the right Is*(err) helper, since the login loop
// branches on those classifications.
func TestExchangeDeviceCode_OAuthErrorMapping(t *testing.T) {
	cases := []struct {
		name              string
		oauthError        string
		wantPending       bool
		wantExpired       bool
		wantAccessDenied  bool
	}{
		{name: "authorization_pending", oauthError: "authorization_pending", wantPending: true},
		{name: "slow_down also pending", oauthError: "slow_down", wantPending: true},
		{name: "expired_token", oauthError: "expired_token", wantExpired: true},
		{name: "access_denied", oauthError: "access_denied", wantAccessDenied: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"` + tc.oauthError + `"}`))
			}))
			t.Cleanup(srv.Close)

			c := New("", WithBaseURL(srv.URL))
			_, err := c.ExchangeDeviceCode(t.Context(), "x")
			require.Error(t, err)
			require.Equal(t, tc.wantPending, IsAuthorizationPending(err), "IsAuthorizationPending mismatch")
			require.Equal(t, tc.wantExpired, IsExpiredToken(err), "IsExpiredToken mismatch")
			require.Equal(t, tc.wantAccessDenied, IsAccessDenied(err), "IsAccessDenied mismatch")
		})
	}
}

func TestExchangeDeviceCode_500FallsThroughToAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	c := New("", WithBaseURL(srv.URL))
	_, err := c.ExchangeDeviceCode(t.Context(), "x")
	require.Error(t, err)
	require.False(t, IsAuthorizationPending(err))
	require.False(t, IsExpiredToken(err))
}
