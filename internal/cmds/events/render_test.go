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

func TestShortID(t *testing.T) {
	require.Equal(t, "01234567", shortID("01234567-89ab-cdef-0123-456789abcdef"))
	require.Equal(t, "short", shortID("short"))
	require.Equal(t, "", shortID(""))
}

func TestRenderEventsList_TableColumnsMatchUI(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, renderEventsList(&buf, "table", []types.Event{sampleEvent()}, true))
	out := buf.String()
	for _, want := range []string{"ID", "TIME", "ACTION", "ACTOR", "TARGET", "TENANT", "RESULT"} {
		require.Containsf(t, out, want, "missing column header %q", want)
	}
	for _, want := range []string{"01234567", "user.login", "alice (user)", "session/s_99", "acme-co", "ok"} {
		require.Containsf(t, out, want, "missing cell value %q", want)
	}
}

func TestRenderEventsList_EmptyTablePrintsMessage(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, renderEventsList(&buf, "table", nil, true))
	require.Contains(t, buf.String(), "No events.")
}

func TestRenderEventsList_HeaderlessForWatch(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, renderEventsList(&buf, "table", []types.Event{sampleEvent()}, false))
	out := buf.String()
	require.NotContains(t, out, "ID   TIME", "watch batches must not re-emit header")
	require.NotContains(t, out, "ACTION", "watch batches must not re-emit header")
}

func TestRenderEventsList_TenantDashWhenEmpty(t *testing.T) {
	e := sampleEvent()
	e.TenantID = ""
	var buf bytes.Buffer
	require.NoError(t, renderEventsList(&buf, "table", []types.Event{e}, true))
	require.Contains(t, buf.String(), "—")
}

// TestColorizeResult covers the plain (non-TTY) path. ANSI wrapping
// itself is exercised in internal/output/color_test.go; here we just
// verify that colorizeResult doesn't mangle status text and returns
// "—" for an empty status. The constructed Stylist is the same one
// production builds when stdout isn't a TTY.
func TestColorizeResult(t *testing.T) {
	for _, status := range []string{"ok", "success", "error", "failure", "denied", "throttled", "weird"} {
		got := colorizeResult(resultSummary{Status: status}, output.PlainStylist())
		require.Equalf(t, status, got, "plain stylist should pass %q through unchanged", status)
	}
	require.Equal(t, "—", colorizeResult(resultSummary{Status: ""}, output.PlainStylist()))
	require.Equal(t, "—", colorizeResult(resultSummary{Status: "  "}, output.PlainStylist()),
		"whitespace-only status should be treated as empty")
}

func TestRenderEventsList_JSONIsBareArray(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, renderEventsList(&buf, "json", []types.Event{sampleEvent()}, true))
	out := strings.TrimSpace(buf.String())
	require.True(t, strings.HasPrefix(out, "["), "list JSON should be a bare array, got:\n%s", out)
	require.Contains(t, out, `"action": "user.login"`)
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

func TestParseTimeFlag(t *testing.T) {
	cases := []struct {
		in        string
		wantZero  bool
		wantRecent bool // true if the result should be "ago" (within last minute)
		wantErr    bool
	}{
		{in: "", wantZero: true},
		{in: "2026-05-09T12:00:00Z"}, // RFC3339
		{in: "1h", wantRecent: true},
		{in: "24h", wantRecent: true},
		{in: "7d", wantRecent: true},
		{in: "0d", wantRecent: true},
		{in: "garbage", wantErr: true},
		{in: "1.5h", wantRecent: true}, // ParseDuration accepts fractional hours
	}
	for _, tc := range cases {
		got, err := parseTimeFlag(tc.in)
		if tc.wantErr {
			require.Errorf(t, err, "input %q", tc.in)
			continue
		}
		require.NoErrorf(t, err, "input %q", tc.in)
		if tc.wantZero {
			require.Truef(t, got.IsZero(), "input %q: want zero", tc.in)
			continue
		}
		require.Falsef(t, got.IsZero(), "input %q: should not be zero", tc.in)
	}
}
