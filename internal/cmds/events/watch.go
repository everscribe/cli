package events

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/config"
	"github.com/everscribe/cli/internal/output"
	"github.com/everscribe/cli/internal/types"
)

// watchPageLimit is the per-poll page cap. Events are typically
// arriving slowly enough that 50 covers a 3-second window in busy
// projects; we bump this internally rather than expose a flag because
// the user-facing knob is --interval.
const watchPageLimit = 50

// watchOverlap is how far back the polling window reaches relative
// to the latest event we've seen. A small overlap shields against
// in-flight events that arrive marginally late and would otherwise
// fall in a "since=last_seen+ε" gap. Dedupe-by-ID handles the rest.
const watchOverlap = 1 * time.Second

func newWatchCmd() *cobra.Command {
	var (
		project  string
		interval time.Duration
		format   string
		filters  filterFlags
	)
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Stream new events by polling, dedupe by event ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWatch(cmd.Context(), cmd.OutOrStdout(), project, filters, interval, format)
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (required)")
	cmd.Flags().DurationVar(&interval, "interval", 3*time.Second, "polling interval")
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	addFilterFlags(cmd, &filters)
	_ = cmd.MarkFlagRequired("project")
	return cmd
}

func runWatch(ctx context.Context, stdout io.Writer, projectID string, ff filterFlags, interval time.Duration, format string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	cf, err := ff.toClientFilter()
	if err != nil {
		return err
	}
	cf.Limit = watchPageLimit

	c := client.New(pat.Token)
	sty := output.NewStylist(stdout)

	// Print column headers up front in table mode so the user sees
	// the layout while waiting. JSON/YAML watchers stream one bare
	// array per batch — easy to parse line-by-line.
	if err := printWatchHeader(stdout, format); err != nil {
		return err
	}

	// Watch starts at "now" — the user just ran `events list` to see
	// history; watch is for tailing.
	highWater := time.Now()
	cf.Since = highWater.Add(-watchOverlap)

	seen := make(map[string]struct{})

	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
		}

		resp, err := c.ListEvents(ctx, projectID, cf)
		if err != nil {
			fmt.Fprintf(stdout, "watch: list failed: %v\n", err)
			continue
		}

		fresh := dedupe(resp.Events, seen)
		if len(fresh) == 0 {
			continue
		}

		// API returns newest-first; reverse so the rolling output
		// reads chronologically (oldest of the batch on top).
		reverse(fresh)

		if err := printWatchBatch(stdout, format, fresh, sty); err != nil {
			return err
		}

		// Advance the polling window. With watchOverlap the next poll
		// re-reads a small slice we've already seen; dedupe filters it.
		latest := fresh[len(fresh)-1].OccurredAt
		if latest.After(highWater) {
			highWater = latest
			cf.Since = highWater.Add(-watchOverlap)
		}
		pruneSeen(seen, fresh)
	}
}

// watchRowFmt is the fixed-width column layout for table-mode watch
// output. tabwriter would re-align columns on each Flush, breaking
// the rolling-append model — fixed widths keep rows visually
// consistent across batches at the cost of occasional truncation.
const watchRowFmt = "%-10s  %-6s  %-25s  %-25s  %-22s  %-12s  %-8s\n"

func printWatchHeader(w io.Writer, format string) error {
	f, err := output.ParseFormat(format, true)
	if err != nil {
		return err
	}
	if f != output.FormatTable {
		return nil
	}
	fmt.Fprintf(w, watchRowFmt, "ID", "TIME", "ACTION", "ACTOR", "TARGET", "TENANT", "RESULT")
	return nil
}

func printWatchBatch(w io.Writer, format string, evs []types.Event, sty output.Stylist) error {
	f, err := output.ParseFormat(format, true)
	if err != nil {
		return err
	}
	switch f {
	case output.FormatJSON:
		return output.JSON(w, evs)
	case output.FormatYAML:
		return output.YAML(w, evs)
	default:
		for _, e := range evs {
			fmt.Fprintf(w, watchRowFmt,
				shortID(e.ID),
				output.Age(e.OccurredAt),
				truncate(e.Action, 25),
				truncate(formatActor(unmarshalActor(e.Actor)), 25),
				truncate(formatTarget(unmarshalTarget(e.Target)), 22),
				truncate(tenantOrDash(e.TenantID), 12),
				colorizeResult(unmarshalResult(e.Result), sty),
			)
		}
		return nil
	}
}

// truncate clips s to max chars, replacing the last with an ellipsis
// when the string was longer. Length is byte-based; UUIDs and ASCII
// action strings are the dominant inputs, so multi-byte handling is
// unnecessary in practice.
func truncate(s string, max int) string {
	if max < 1 || len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func dedupe(evs []types.Event, seen map[string]struct{}) []types.Event {
	out := evs[:0:0]
	for _, e := range evs {
		if _, ok := seen[e.ID]; ok {
			continue
		}
		seen[e.ID] = struct{}{}
		out = append(out, e)
	}
	return out
}

func reverse(evs []types.Event) {
	for i, j := 0, len(evs)-1; i < j; i, j = i+1, j-1 {
		evs[i], evs[j] = evs[j], evs[i]
	}
}

// pruneSeen guards against unbounded memory growth on long-running
// watch sessions. We only need entries within the current polling
// window — anything older than `since` can't reappear. As a
// heuristic, when seen grows past maxSeen we drop everything and
// repopulate from the current batch; the overlap window will refill
// it on the next tick. The worst case is one batch of duplicates
// after a clear, which the user won't notice.
func pruneSeen(seen map[string]struct{}, batch []types.Event) {
	const maxSeen = 10000
	if len(seen) <= maxSeen {
		return
	}
	for k := range seen {
		delete(seen, k)
	}
	for _, e := range batch {
		seen[e.ID] = struct{}{}
	}
}
