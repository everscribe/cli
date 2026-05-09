package types

import "time"

// IssueDeviceCodeResponse is the body of POST /v1/cli/device-codes.
// Field names mirror RFC 8628 (OAuth 2.0 Device Authorization Grant).
type IssueDeviceCodeResponse struct {
	UserCode                string `json:"user_code"`
	DeviceCode              string `json:"device_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"` // seconds
	Interval                int    `json:"interval"`   // suggested polling interval, seconds
}

// ExchangeDeviceCodeRequest is the body of POST /v1/cli/device-tokens.
type ExchangeDeviceCodeRequest struct {
	DeviceCode string `json:"device_code"`
}

// ExchangeDeviceCodeResponse is the success body of the exchange
// endpoint. PATExpiresAt is the zero value when the PAT has no expiry.
type ExchangeDeviceCodeResponse struct {
	Plaintext    string    `json:"plaintext"`
	PATID        string    `json:"pat_id"`
	UserID       string    `json:"user_id"`
	UserEmail    string    `json:"user_email"`
	PATExpiresAt time.Time `json:"pat_expires_at,omitempty"`
}

// RFC 8628 error codes the exchange endpoint returns with a 400 status.
// Login polling matches on these to decide whether to retry or abort.
const (
	DeviceErrorAuthorizationPending = "authorization_pending"
	DeviceErrorSlowDown             = "slow_down"
	DeviceErrorAccessDenied         = "access_denied"
	DeviceErrorExpiredToken         = "expired_token"
)
