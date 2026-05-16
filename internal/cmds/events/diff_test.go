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

// TestRenderDiff covers the rendered output for normal updates,
// create-events (before=null), delete-events (after=null), and the
// two zero-input edge cases that must surface as ErrNoChange.
func TestRenderDiff(t *testing.T) {
	cases := []struct {
		name       string
		change     json.RawMessage
		wantErr    error  // non-nil → assert errors.Is(err, wantErr) and skip output checks
		wantErrSub string // non-nil → assert ErrorContains and skip output checks
		wantSub    []string
	}{
		{
			name: "happy path: status change",
			change: json.RawMessage(`{
				"before": {"name": "ingest", "status": "active", "owner": "alice"},
				"after":  {"name": "ingest", "status": "paused", "owner": "alice"}
			}`),
			wantSub: []string{
				"--- before",
				"+++ after",
				`-  "status": "active"`,
				`+  "status": "paused"`,
				` "name": "ingest"`,
			},
		},
		{
			name: "create event: before is null",
			change: json.RawMessage(`{
				"before": null,
				"after":  {"name": "new-thing"}
			}`),
			wantSub: []string{"-null", `+  "name": "new-thing"`},
		},
		{
			name: "delete event: after is null",
			change: json.RawMessage(`{
				"before": {"name": "old-thing"},
				"after":  null
			}`),
			wantSub: []string{`-  "name": "old-thing"`, "+null"},
		},
		{
			name:    "empty raw is ErrNoChange",
			change:  nil,
			wantErr: ErrNoChange,
		},
		{
			name:    "both before and after empty is ErrNoChange",
			change:  json.RawMessage(`{}`),
			wantErr: ErrNoChange,
		},
		{
			name:       "malformed change surfaces parse error",
			change:     json.RawMessage(`not json`),
			wantErrSub: "parse change",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := renderDiff(&buf, tc.change, output.PlainStylist())
			switch {
			case tc.wantErr != nil:
				require.ErrorIs(t, err, tc.wantErr)
				return
			case tc.wantErrSub != "":
				require.Error(t, err)
				require.False(t, errors.Is(err, ErrNoChange))
				require.ErrorContains(t, err, tc.wantErrSub)
				return
			}
			require.NoError(t, err)
			out := buf.String()
			for _, want := range tc.wantSub {
				require.Containsf(t, out, want, "missing %q", want)
			}
		})
	}
}

func TestPrettyJSON(t *testing.T) {
	cases := []struct {
		name string
		in   json.RawMessage
		want string
	}{
		{name: "nil renders null", in: nil, want: "null"},
		{name: "literal null renders null", in: json.RawMessage("null"), want: "null"},
		{name: "object pretty-prints", in: json.RawMessage(`{"a":1}`), want: "{\n  \"a\": 1\n}"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, prettyJSON(tc.in))
		})
	}
}

func kindPtr(k diffKind) *diffKind { return &k }

// TestLineDiff covers the bulk shapes the LCS-driven diff must
// produce. The "middle changed" case allows either order for the
// added/removed pair since the LCS algorithm may pick either.
func TestLineDiff(t *testing.T) {
	cases := []struct {
		name        string
		before      []string
		after       []string
		wantLen     int
		wantAllKind *diffKind // when non-nil, every entry must have this kind
		extra       func(t *testing.T, got []diffEntry)
	}{
		{
			name:        "all same",
			before:      []string{"a", "b", "c"},
			after:       []string{"a", "b", "c"},
			wantLen:     3,
			wantAllKind: kindPtr(diffSame),
		},
		{
			name:        "all added",
			before:      nil,
			after:       []string{"a", "b"},
			wantLen:     2,
			wantAllKind: kindPtr(diffAdded),
		},
		{
			name:        "all removed",
			before:      []string{"a", "b"},
			after:       nil,
			wantLen:     2,
			wantAllKind: kindPtr(diffRemoved),
		},
		{
			name:    "middle changed",
			before:  []string{"a", "b", "c"},
			after:   []string{"a", "X", "c"},
			wantLen: 4,
			extra: func(t *testing.T, got []diffEntry) {
				require.Equal(t, diffSame, got[0].kind)
				require.Equal(t, "a", got[0].text)
				require.Equal(t, "c", got[3].text)
				gotKinds := strings.Join([]string{
					string(rune('0' + int(got[1].kind))),
					string(rune('0' + int(got[2].kind))),
				}, "")
				gotTexts := got[1].text + got[2].text
				require.Contains(t, "12 21", gotKinds, "middle should be one added + one removed")
				require.Contains(t, gotTexts, "b")
				require.Contains(t, gotTexts, "X")
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := lineDiff(tc.before, tc.after)
			require.Len(t, got, tc.wantLen)
			if tc.wantAllKind != nil {
				for _, e := range got {
					require.Equal(t, *tc.wantAllKind, e.kind)
				}
			}
			if tc.extra != nil {
				tc.extra(t, got)
			}
		})
	}
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
