package events

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/config"
	"github.com/everscribe/cli/internal/output"
	"github.com/everscribe/cli/internal/types"
)

func newListCmd() *cobra.Command {
	var (
		project string
		limit   int
		all     bool
		format  string
		prompt  string
		query   string
		filters filterFlags
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List events in a project (most recent first)",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := config.ResolveProjectID(project)
			if err != nil {
				return err
			}
			return runList(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), projectID, filters, prompt, query, limit, all, format)
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (defaults to the project saved by 'es projects use')")
	cmd.Flags().IntVar(&limit, "limit", 50, "max events per page")
	cmd.Flags().BoolVar(&all, "all", false, "fetch all pages instead of one")
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	cmd.Flags().StringVar(&prompt, "prompt", "", "natural-language query; translated to a filter DSL by the API (mutually exclusive with --query and the structured filter flags)")
	cmd.Flags().StringVar(&query, "query", "", "structured filter DSL passed directly to the API (mutually exclusive with --prompt and the structured filter flags)")
	addFilterFlags(cmd, &filters)
	return cmd
}

func runList(ctx context.Context, stdout, stderr io.Writer, projectID string, ff filterFlags, prompt, query string, limit int, all bool, format string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	if err := validateQueryFlags(ff, prompt, query); err != nil {
		return err
	}
	cf, err := ff.toClientFilter()
	if err != nil {
		return err
	}
	cf.Limit = limit

	c := client.New(pat.Token)

	switch {
	case prompt != "":
		nlp, err := c.GenerateNLPFilters(ctx, projectID, prompt)
		if err != nil {
			return fmt.Errorf("translate prompt: %w", err)
		}
		if nlp.DSL == "" {
			return promptTranslationError(prompt, nlp)
		}
		cf.Q = nlp.DSL
		if parsedFormat(format) == output.FormatTable {
			printPromptTranslation(stderr, nlp)
		}
	case query != "":
		cf.Q = query
	}

	var allEvents []types.Event
	resp, err := c.ListEvents(ctx, projectID, cf)
	if err != nil {
		return err
	}
	allEvents = resp.Events

	if all {
		for resp.NextCursor != "" {
			cf.Cursor = resp.NextCursor
			resp, err = c.ListEvents(ctx, projectID, cf)
			if err != nil {
				return err
			}
			allEvents = append(allEvents, resp.Events...)
		}
	}

	if err := renderEventsList(stdout, format, allEvents, true); err != nil {
		return err
	}

	// Pagination hint is only useful in human (table) output. Skip
	// for json/yaml so machine consumers don't get a stray comment
	// mid-stream.
	if !all && resp.NextCursor != "" && parsedFormat(format) == output.FormatTable {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "More events available. Pass --all to fetch every page.")
	}
	return nil
}

// validateQueryFlags enforces the mutual exclusion between --prompt,
// --query, and the structured filter flags. The CLI rejects mixed
// usage even though the API would technically merge flat params with
// a DSL — combining a translated NL prompt with hand-written filters
// makes the resulting query opaque to users, so we force one mode.
func validateQueryFlags(ff filterFlags, prompt, query string) error {
	if prompt != "" && query != "" {
		return fmt.Errorf("--prompt and --query are mutually exclusive")
	}
	if prompt == "" && query == "" {
		return nil
	}
	if structured := setStructuredFilters(ff); len(structured) > 0 {
		mode := "--prompt"
		if query != "" {
			mode = "--query"
		}
		return fmt.Errorf("%s cannot be combined with structured filter flags: %v", mode, structured)
	}
	return nil
}

// setStructuredFilters returns the names of the filterFlags fields that
// the user populated, used to build a precise mutex error message.
func setStructuredFilters(ff filterFlags) []string {
	var names []string
	for _, f := range []struct {
		name, val string
	}{
		{"--since", ff.since},
		{"--before", ff.before},
		{"--action", ff.action},
		{"--actor", ff.actor},
		{"--actor-type", ff.actorType},
		{"--target-type", ff.targetType},
		{"--tenant", ff.tenant},
	} {
		if f.val != "" {
			names = append(names, f.name)
		}
	}
	return names
}

// promptTranslationError builds a user-facing error when the NLP
// endpoint succeeded but couldn't produce a DSL. We refuse to fall
// through to an unfiltered list, since that would silently return
// every event in the project.
func promptTranslationError(prompt string, nlp *types.GenerateNLPFiltersResponse) error {
	msg := fmt.Sprintf("could not translate prompt %q to a filter", prompt)
	if nlp.Explanation != "" {
		msg += ": " + nlp.Explanation
	}
	if len(nlp.Unsupported) > 0 {
		msg += fmt.Sprintf(" (unsupported: %v)", nlp.Unsupported)
	}
	return fmt.Errorf("%s", msg)
}

// printPromptTranslation surfaces the translated DSL on stderr so
// table-mode users can see what the NLP layer produced — both for
// transparency and to teach the DSL syntax.
func printPromptTranslation(stderr io.Writer, nlp *types.GenerateNLPFiltersResponse) {
	fmt.Fprintf(stderr, "→ translated to DSL: %s\n", nlp.DSL)
	if len(nlp.Unsupported) > 0 {
		fmt.Fprintf(stderr, "  unsupported parts ignored: %v\n", nlp.Unsupported)
	}
}

// parsedFormat is a best-effort re-parse of --format that returns the
// zero value for invalid input — used only by display-only branches
// where the renderer has already validated the flag.
func parsedFormat(s string) output.Format {
	f, _ := output.ParseFormat(s, true)
	return f
}
