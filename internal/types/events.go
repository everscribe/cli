package types

import (
	"encoding/json"
	"time"
)

// Event mirrors the API/SDK on-the-wire shape. The nested actor /
// target / metadata / origin / result / change fields are kept as
// raw JSON because their schemas are application-defined per ingest
// (the customer chooses what shape "actor" takes, etc.).
type Event struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenant_id,omitempty"`
	OccurredAt     time.Time       `json:"occurred_at"`
	Actor          json.RawMessage `json:"actor"`
	Action         string          `json:"action"`
	Target         json.RawMessage `json:"target,omitempty"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
	Origin         json.RawMessage `json:"origin,omitempty"`
	Result         json.RawMessage `json:"result,omitempty"`
	Change         json.RawMessage `json:"change,omitempty"`
	IdempotencyKey string          `json:"idempotency_key,omitempty"`
}

type ListEventsResponse struct {
	Events     []Event `json:"events"`
	NextCursor string  `json:"next_cursor,omitempty"`
}

type EventResponse struct {
	Event Event `json:"event"`
}
