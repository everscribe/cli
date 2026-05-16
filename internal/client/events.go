package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/everscribe/cli/internal/types"
)

// ListEventsFilter is the query-parameter set the events list endpoint
// accepts. Zero-valued fields are omitted from the request.
//
// Q carries the structured filter DSL passed through `?q=...`. The
// server merges it with the flat fields (Since, Action, ...) but the
// CLI sends one or the other — never both — to keep semantics obvious
// to the user.
type ListEventsFilter struct {
	Since      time.Time
	Before     time.Time
	Cursor     string
	Limit      int
	Action     string
	Actor      string
	ActorType  string
	TenantID   string
	TargetType string
	Q          string
}

// query renders the filter as URL values.
func (f ListEventsFilter) query() url.Values {
	q := url.Values{}
	if !f.Since.IsZero() {
		q.Set("since", f.Since.UTC().Format(time.RFC3339))
	}
	if !f.Before.IsZero() {
		q.Set("before", f.Before.UTC().Format(time.RFC3339))
	}
	if f.Cursor != "" {
		q.Set("cursor", f.Cursor)
	}
	if f.Limit > 0 {
		q.Set("limit", strconv.Itoa(f.Limit))
	}
	if f.Action != "" {
		q.Set("action", f.Action)
	}
	if f.Actor != "" {
		q.Set("actor", f.Actor)
	}
	if f.ActorType != "" {
		q.Set("actor_type", f.ActorType)
	}
	if f.TenantID != "" {
		q.Set("tenant_id", f.TenantID)
	}
	if f.TargetType != "" {
		q.Set("target_type", f.TargetType)
	}
	if f.Q != "" {
		q.Set("q", f.Q)
	}
	return q
}

func (c *Client) ListEvents(ctx context.Context, projectID string, f ListEventsFilter) (*types.ListEventsResponse, error) {
	var out types.ListEventsResponse
	if err := c.Do(ctx, http.MethodGet, "/v1/projects/"+projectID+"/events", f.query(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GenerateNLPFilters translates a natural-language query into the
// events filter DSL by calling POST /v1/projects/{projectID}/events/nlp.
// The returned DSL can then be passed to ListEvents via ListEventsFilter.Q.
func (c *Client) GenerateNLPFilters(ctx context.Context, projectID, prompt string) (*types.GenerateNLPFiltersResponse, error) {
	var out types.GenerateNLPFiltersResponse
	body := types.GenerateNLPFiltersRequest{Q: prompt}
	if err := c.Do(ctx, http.MethodPost, "/v1/projects/"+projectID+"/events/nlp", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetEvent(ctx context.Context, projectID, eventID string) (*types.Event, error) {
	var out types.EventResponse
	if err := c.Do(ctx, http.MethodGet, "/v1/projects/"+projectID+"/events/"+eventID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out.Event, nil
}
