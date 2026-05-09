package events

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/everscribe/cli/internal/output"
)

func TestRenderDiff_HappyPath(t *testing.T) {
	change := json.RawMessage(`{
		"before": {"name": "ingest", "status": "active", "owner": "alice"},
		"after":  {"name": "ingest", "status": "paused", "owner": "alice"}
	}`)

	var buf bytes.Buffer
	require.NoError(t, renderDiff(&buf, change, output.PlainStylist()))

	out := buf.String()
	require.Contains(t, out, "--- before")
	require.Contains(t, out, "+++ after")
	require.Contains(t, out, `-  "status": "active"`, "removed line should be marked with -")
	require.Contains(t, out, `+  "status": "paused"`, "added line should be marked with +")
	require.Contains(t, out, ` "name": "ingest"`, "unchanged line should have leading space")
}

func TestRenderDiff_EmptyRawReturnsErrNoChange(t *testing.T) {
	err := renderDiff(&bytes.Buffer{}, nil, output.PlainStylist())
	require.ErrorIs(t, err, ErrNoChange)
}

func TestRenderDiff_BothBeforeAndAfterEmptyReturnsErrNoChange(t *testing.T) {
	change := json.RawMessage(`{}`)
	err := renderDiff(&bytes.Buffer{}, change, output.PlainStylist())
	require.ErrorIs(t, err, ErrNoChange)
}

func TestRenderDiff_CreateEvent_BeforeIsNull(t *testing.T) {
	change := json.RawMessage(`{
		"before": null,
		"after":  {"name": "new-thing"}
	}`)
	var buf bytes.Buffer
	require.NoError(t, renderDiff(&buf, change, output.PlainStylist()))
	out := buf.String()
	require.Contains(t, out, "-null", "create event should diff null before vs new after")
	require.Contains(t, out, `+  "name": "new-thing"`)
}

func TestRenderDiff_DeleteEvent_AfterIsNull(t *testing.T) {
	change := json.RawMessage(`{
		"before": {"name": "old-thing"},
		"after":  null
	}`)
	var buf bytes.Buffer
	require.NoError(t, renderDiff(&buf, change, output.PlainStylist()))
	out := buf.String()
	require.Contains(t, out, `-  "name": "old-thing"`)
	require.Contains(t, out, "+null")
}

func TestRenderDiff_MalformedChangeReturnsParseError(t *testing.T) {
	err := renderDiff(&bytes.Buffer{}, json.RawMessage(`not json`), output.PlainStylist())
	require.Error(t, err)
	require.False(t, errors.Is(err, ErrNoChange))
	require.Contains(t, err.Error(), "parse change")
}

func TestPrettyJSON(t *testing.T) {
	require.Equal(t, "null", prettyJSON(nil))
	require.Equal(t, "null", prettyJSON(json.RawMessage("null")))
	require.Equal(t, "{\n  \"a\": 1\n}", prettyJSON(json.RawMessage(`{"a":1}`)))
}

func TestLineDiff_AllSame(t *testing.T) {
	got := lineDiff([]string{"a", "b", "c"}, []string{"a", "b", "c"})
	require.Len(t, got, 3)
	for _, e := range got {
		require.Equal(t, diffSame, e.kind)
	}
}

func TestLineDiff_AllAdded(t *testing.T) {
	got := lineDiff(nil, []string{"a", "b"})
	require.Len(t, got, 2)
	for _, e := range got {
		require.Equal(t, diffAdded, e.kind)
	}
}

func TestLineDiff_AllRemoved(t *testing.T) {
	got := lineDiff([]string{"a", "b"}, nil)
	require.Len(t, got, 2)
	for _, e := range got {
		require.Equal(t, diffRemoved, e.kind)
	}
}

func TestLineDiff_MiddleChanged(t *testing.T) {
	got := lineDiff([]string{"a", "b", "c"}, []string{"a", "X", "c"})
	// Expected: same(a), removed(b), added(X), same(c)
	require.Len(t, got, 4)
	require.Equal(t, diffSame, got[0].kind)
	require.Equal(t, "a", got[0].text)
	require.Equal(t, "c", got[3].text)
	// Middle: one removed and one added in some order — LCS may pick
	// either order but both must be present.
	gotKinds := strings.Join([]string{
		string(rune('0' + int(got[1].kind))),
		string(rune('0' + int(got[2].kind))),
	}, "")
	gotTexts := got[1].text + got[2].text
	require.Contains(t, "12 21", gotKinds, "middle should be one added + one removed")
	require.Contains(t, gotTexts, "b")
	require.Contains(t, gotTexts, "X")
}

func TestRenderDiff_TTYColorWraps(t *testing.T) {
	// Construct a stylist with color forced on by going through the
	// constructor (we can't directly, since ColorEnabled checks for a
	// TTY). Verify that when we pass a *bytes.Buffer (not TTY), no
	// ANSI sequences appear — proves the color path is gated.
	change := json.RawMessage(`{"before":{"x":1}, "after":{"x":2}}`)
	var buf bytes.Buffer
	require.NoError(t, renderDiff(&buf, change, output.NewStylist(&buf)))
	require.NotContains(t, buf.String(), "\x1b[", "non-TTY output must not contain ANSI escapes")
}
