package types

import "time"

// APIKey is a project-scoped ingest credential. Plaintext is only
// returned by the create endpoint via CreateAPIKeyResponse — list
// responses redact to the prefix.
type APIKey struct {
	ID         string    `json:"id" yaml:"id"`
	ProjectID  string    `json:"project_id" yaml:"project_id"`
	Name       string    `json:"name" yaml:"name"`
	Prefix     string    `json:"prefix" yaml:"prefix"`
	LastUsedAt time.Time `json:"last_used_at,omitempty" yaml:"last_used_at,omitempty"`
	RevokedAt  time.Time `json:"revoked_at,omitempty" yaml:"revoked_at,omitempty"`
	CreatedAt  time.Time `json:"created_at" yaml:"created_at"`
}

type CreateAPIKeyRequest struct {
	Name string `json:"name"`
}

type CreateAPIKeyResponse struct {
	Key       APIKey `json:"key" yaml:"key"`
	Plaintext string `json:"plaintext" yaml:"plaintext"`
}

type ListAPIKeysResponse struct {
	Keys []APIKey `json:"keys"`
}
