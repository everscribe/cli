// Package events implements `es events ...`.
package events

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
)

// NewCmd returns the `es events` command tree.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "events",
		Aliases: []string{"event"},
		Short:   "List, watch, describe, and diff project events",
	}
	cmd.AddCommand(
		newListCmd(),
		newWatchCmd(),
		newDescribeCmd(),
		newDiffCmd(),
	)
	return cmd
}

// filterFlags is the shared filter set for `events list` and `events watch`.
type filterFlags struct {
	since      string
	before     string
	action     string
	actor      string
	actorType  string
	targetType string
	tenant     string
}

// addFilterFlags attaches the shared filter flags to cmd. Each individual
// command also defines its own command-specific flags (limit, interval, etc.).
func addFilterFlags(cmd *cobra.Command, f *filterFlags) {
	cmd.Flags().StringVar(&f.since, "since", "", "include only events at or after this time (RFC3339 or duration like 1h, 7d)")
	cmd.Flags().StringVar(&f.before, "before", "", "include only events strictly before this time (RFC3339 or duration like 1h, 7d)")
	cmd.Flags().StringVar(&f.action, "action", "", "filter by action string (e.g. user.login)")
	cmd.Flags().StringVar(&f.actor, "actor", "", "filter by actor ID")
	cmd.Flags().StringVar(&f.actorType, "actor-type", "", "filter by actor type (e.g. user, service)")
	cmd.Flags().StringVar(&f.targetType, "target-type", "", "filter by target type")
	cmd.Flags().StringVar(&f.tenant, "tenant", "", "filter by tenant ID")
}

// toClientFilter converts the cobra-bound flag struct into the API
// client's filter, parsing the time strings as it goes.
func (f filterFlags) toClientFilter() (client.ListEventsFilter, error) {
	since, err := parseTimeFlag(f.since)
	if err != nil {
		return client.ListEventsFilter{}, fmt.Errorf("--since: %w", err)
	}
	before, err := parseTimeFlag(f.before)
	if err != nil {
		return client.ListEventsFilter{}, fmt.Errorf("--before: %w", err)
	}
	return client.ListEventsFilter{
		Since:      since,
		Before:     before,
		Action:     f.action,
		Actor:      f.actor,
		ActorType:  f.actorType,
		TenantID:   f.tenant,
		TargetType: f.targetType,
	}, nil
}

// parseTimeFlag accepts an RFC3339 timestamp or a duration ("1h",
// "24h", "7d"). Durations are interpreted as "<that long> ago".
// Returns the zero time for an empty string.
func parseTimeFlag(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if d, err := time.ParseDuration(s); err == nil {
		return time.Now().Add(-d), nil
	}
	// Go's time.ParseDuration doesn't handle "d"; we extend it.
	if strings.HasSuffix(s, "d") {
		if days, err := strconv.Atoi(strings.TrimSuffix(s, "d")); err == nil && days >= 0 {
			return time.Now().Add(-time.Duration(days) * 24 * time.Hour), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time %q (use RFC3339 like 2026-05-09T12:00:00Z or duration like 1h, 24h, 7d)", s)
}
