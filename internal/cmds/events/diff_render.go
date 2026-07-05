package events

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/everscribe/cli/internal/output"
)

// decodeRawJSON unmarshals a json.RawMessage into a generic any tree.
// Returns nil for empty input so YAML/JSON output omits the field
// (omitempty on the host struct). Unparseable JSON returns the raw
// string so the user still sees the data.
func decodeRawJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	return v
}

// renderDiff writes a unified diff of change.before vs change.after.
// Returns ErrNoChange if the event has no Change field (most events
// won't), or if the Change exists but neither before nor after is set.
//
// Strategy mirrors the Everscribe server's diff renderer: pretty-print
// each side as indented JSON, line-split, run an LCS-based line diff,
// and emit `+`/`-` lines with optional color when stdout is a TTY.
func renderDiff(w io.Writer, raw json.RawMessage, sty output.Stylist) error {
	if len(raw) == 0 {
		return ErrNoChange
	}
	var change struct {
		Before json.RawMessage `json:"before"`
		After  json.RawMessage `json:"after"`
	}
	if err := json.Unmarshal(raw, &change); err != nil {
		return fmt.Errorf("parse change: %w", err)
	}
	if len(change.Before) == 0 && len(change.After) == 0 {
		return ErrNoChange
	}

	beforeLines := strings.Split(prettyJSON(change.Before), "\n")
	afterLines := strings.Split(prettyJSON(change.After), "\n")
	hunk := lineDiff(beforeLines, afterLines)

	fmt.Fprintln(w, sty.Red("--- before"))
	fmt.Fprintln(w, sty.Green("+++ after"))
	for _, ln := range hunk {
		switch ln.kind {
		case diffSame:
			fmt.Fprintln(w, " "+ln.text)
		case diffRemoved:
			fmt.Fprintln(w, sty.Red("-"+ln.text))
		case diffAdded:
			fmt.Fprintln(w, sty.Green("+"+ln.text))
		}
	}
	return nil
}

// ErrNoChange is returned by renderDiff when there's nothing to diff.
// Surfaced through to runDiff so the user gets a clear message.
var ErrNoChange = fmt.Errorf("event has no change record")

// prettyJSON re-renders raw JSON with 2-space indentation. Returns
// "null" for empty/null input - the diff still has something to
// show against (create events have before=null; delete events have
// after=null).
func prettyJSON(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return "null"
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return strings.TrimSpace(string(raw))
	}
	buf, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return strings.TrimSpace(string(raw))
	}
	return string(buf)
}

type diffKind int

const (
	diffSame diffKind = iota
	diffAdded
	diffRemoved
)

type diffEntry struct {
	kind diffKind
	text string
}

// lineDiff computes a unified-style diff via LCS. Mirrors the algorithm
// the Everscribe server uses; the only difference is the output
// shape (linear list of {kind, text} entries instead of side-by-side
// pairs). For event diffs the inputs are small (under 100 lines per
// side typically) so the O(n*m) DP is fine.
func lineDiff(before, after []string) []diffEntry {
	n, m := len(before), len(after)
	if n == 0 && m == 0 {
		return nil
	}

	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if before[i] == after[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	out := make([]diffEntry, 0, n+m)
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case before[i] == after[j]:
			out = append(out, diffEntry{kind: diffSame, text: before[i]})
			i++
			j++
		case dp[i+1][j] >= dp[i][j+1]:
			out = append(out, diffEntry{kind: diffRemoved, text: before[i]})
			i++
		default:
			out = append(out, diffEntry{kind: diffAdded, text: after[j]})
			j++
		}
	}
	for ; i < n; i++ {
		out = append(out, diffEntry{kind: diffRemoved, text: before[i]})
	}
	for ; j < m; j++ {
		out = append(out, diffEntry{kind: diffAdded, text: after[j]})
	}
	return out
}
