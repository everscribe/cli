package events

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/everscribe/cli/internal/output"
	"github.com/everscribe/cli/internal/types"
)

func sampleEvent() types.Event {
	return types.Event{
		ID:         "01234567-89ab-cdef-0123-456789abcdef",
		TenantID:   "acme-co",
		OccurredAt: time.Now().Add(-2 * time.Minute),
		Action:     "user.login",
		Actor:      json.RawMessage(`{"type":"user","id":"u_1","display_name":"alice"}`),
		Target:     json.RawMessage(`{"type":"session","id":"s_99"}`),
		Result:     json.RawMessage(`{"status":"ok"}`),
	}
}

// TestRenderEventsList covers the table/json paths together with
// header-on-watch behavior and the empty-tenant dash. mutate runs on
// the fixture so individual cases don't have to copy the helper.
func TestRenderEventsList(t *testing.T) {
	cases := []struct {
		name       string
		format     string
		mutate     func(*types.Event)
		empty      bool
		header     bool
		wantSub    []string
		wantNotSub []string
	}{
		{
			name:   "table columns and cells match UI",
			format: "table",
			header: true,
			wantSub: []string{
				"ID", "TIME", "ACTION", "ACTOR", "TARGET", "TENANT", "RESULT",
				"01234567-89ab-cdef-0123-456789abcdef", // full UUID, not truncated
				"user.login", "alice (user)", "session/s_99", "acme-co", "ok",
			},
		},
		{
			name:    "empty table prints message",
			format:  "table",
			empty:   true,
			header:  true,
			wantSub: []string{"No events."},
		},
		{
			name:       "headerless watch batches",
			format:     "table",
			header:     false,
			wantNotSub: []string{"ID   TIME", "ACTION"},
		},
		{
			name:    "tenant dash when empty",
			format:  "table",
			header:  true,
			mutate:  func(e *types.Event) { e.TenantID = "" },
			wantSub: []string{"—"},
		},
		{
			name:    "json is bare array",
			format:  "json",
			header:  true,
			wantSub: []string{`"action": "user.login"`},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var events []types.Event
			if !tc.empty {
				e := sampleEvent()
				if tc.mutate != nil {
					tc.mutate(&e)
				}
				events = []types.Event{e}
			}

			var buf bytes.Buffer
			require.NoError(t, renderEventsList(&buf, tc.format, events, tc.header))
			out := buf.String()

			if tc.format == "json" && len(events) > 0 {
				require.True(t, strings.HasPrefix(strings.TrimSpace(out), "["),
					"list JSON should be a bare array, got:\n%s", out)
			}
			for _, want := range tc.wantSub {
				require.Containsf(t, out, want, "missing %q", want)
			}
			for _, notWant := range tc.wantNotSub {
				require.NotContainsf(t, out, notWant, "watch batches must not re-emit header %q", notWant)
			}
		})
	}
}

// TestColorizeResult covers the plain (non-TTY) path. ANSI wrapping
// itself is exercised in internal/output/color_test.go; here we just
// verify that colorizeResult doesn't mangle status text and returns
// "—" for empty/whitespace-only input.
func TestColorizeResult(t *testing.T) {
	cases := []struct {
		name   string
		status string
		want   string
	}{
		{name: "ok passes through", status: "ok", want: "ok"},
		{name: "success passes through", status: "success", want: "success"},
		{name: "error passes through", status: "error", want: "error"},
		{name: "failure passes through", status: "failure", want: "failure"},
		{name: "denied passes through", status: "denied", want: "denied"},
		{name: "throttled passes through", status: "throttled", want: "throttled"},
		{name: "unknown status passes through", status: "weird", want: "weird"},
		{name: "empty status renders dash", status: "", want: "—"},
		{name: "whitespace status renders dash", status: "  ", want: "—"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, colorizeResult(resultSummary{Status: tc.status}, output.PlainStylist()))
		})
	}
}

func TestBuildDescribeView_DecodesNestedFields(t *testing.T) {
	e := sampleEvent()
	view := buildDescribeView(e)

	require.Equal(t, "01234567-89ab-cdef-0123-456789abcdef", view.ID)
	require.Equal(t, "user.login", view.Action)
	require.Equal(t, "acme-co", view.TenantID)

	actor, ok := view.Actor.(map[string]any)
	require.True(t, ok, "actor should decode to a map, got %T", view.Actor)
	require.Equal(t, "alice", actor["display_name"])
	require.Equal(t, "u_1", actor["id"])
}

func TestBuildDescribeView_OmitsEmptyOptionalFields(t *testing.T) {
	e := types.Event{
		ID:         "id-1",
		OccurredAt: time.Now(),
		Action:     "x.y",
	}
	view := buildDescribeView(e)
	require.Nil(t, view.Actor)
	require.Nil(t, view.Target)
	require.Nil(t, view.Metadata)
	require.Nil(t, view.Change)

	var buf bytes.Buffer
	require.NoError(t, output.YAML(&buf, view))
	require.NotContains(t, buf.String(), "actor:", "empty actor should be omitted")
}

// TestParseTimeFlag covers the three accepted shapes (RFC3339,
// Go-style duration, "Nd" extension) plus the empty/invalid edges.
func TestParseTimeFlag(t *testing.T) {
	cases := []struct {
		name       string
		in         string
		wantZero   bool
		wantRecent bool // result should be non-zero (within last minute)
		wantErr    bool
	}{
		{name: "empty", in: "", wantZero: true},
		{name: "rfc3339", in: "2026-05-09T12:00:00Z"},
		{name: "1h duration", in: "1h", wantRecent: true},
		{name: "24h duration", in: "24h", wantRecent: true},
		{name: "7d extension", in: "7d", wantRecent: true},
		{name: "0d extension", in: "0d", wantRecent: true},
		{name: "fractional hours", in: "1.5h", wantRecent: true},
		{name: "garbage rejected", in: "garbage", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseTimeFlag(tc.in)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tc.wantZero {
				require.True(t, got.IsZero(), "want zero")
				return
			}
			require.False(t, got.IsZero(), "should not be zero")
		})
	}
}
