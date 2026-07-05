package skills

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/output"
	"github.com/everscribe/cli/internal/types"
)

// describeView is what gets emitted in JSON / YAML - the catalog
// entry plus the fetched SKILL.md body inline. Tag names match the
// catalog's JSON field names so `es skills describe setup -o json
// | jq '.body'` (etc.) feels consistent across surfaces.
type describeView struct {
	types.Skill
	Body string `json:"body" yaml:"body"`
}

func newDescribeCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "describe <skill>",
		Short: "Show a skill's metadata and SKILL.md body without installing",
		Long: "Fetches the catalog entry for the named skill plus the raw SKILL.md\n" +
			"body. Useful for previewing what `es skills download <name>` would\n" +
			"write to your Claude Code skills directory.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDescribe(cmd.Context(), cmd.OutOrStdout(), args[0], format)
		},
	}
	cmd.Flags().StringVar(&format, "format", "markdown", "output format: markdown | json | yaml")
	return cmd
}

func runDescribe(ctx context.Context, stdout io.Writer, name, format string) error {
	f := NewFetcher()
	cat, err := f.Catalog(ctx)
	if err != nil {
		return err
	}
	skill, ok := findSkill(cat.Skills, name)
	if !ok {
		return fmt.Errorf("skill %q not found in catalog (try `es skills list`)", name)
	}
	bodyBytes, err := f.SkillBody(ctx, skill.Source)
	if err != nil {
		return err
	}
	view := describeView{Skill: skill, Body: string(bodyBytes)}

	switch strings.ToLower(format) {
	case "json":
		return output.JSON(stdout, view)
	case "yaml":
		return output.YAML(stdout, view)
	case "markdown", "":
		return renderSkillMarkdown(stdout, view)
	default:
		return fmt.Errorf("--format=%s invalid (must be markdown, json, or yaml)", format)
	}
}

// renderSkillMarkdown is the default human-friendly format: a small
// key/value metadata block followed by a separator and the raw
// SKILL.md body. yaml-encoded markdown is technically structured but
// painful to scan, so describe defaults to "show me the markdown."
func renderSkillMarkdown(w io.Writer, v describeView) error {
	if _, err := fmt.Fprintf(w,
		"Name:        %s\nVersion:     %s\nDescription: %s\nSkill dir:   %s\nSource:      %s\n\n---\n\n",
		v.Name, v.Version, v.Description, v.ClaudeCodeSkillName, v.Source,
	); err != nil {
		return err
	}
	body := v.Body
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	_, err := fmt.Fprint(w, body)
	return err
}
