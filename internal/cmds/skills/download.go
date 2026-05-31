package skills

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/types"
)

const skillFileMode = 0o644
const skillDirMode = 0o755

// DefaultSkillsDirEnv lets tests (and unusual setups) redirect the
// download target. When unset, downloads land under
// ~/.claude/skills/<claude_code_skill_name>/SKILL.md, which is the
// path Claude Code reads on startup to discover skills.
const DefaultSkillsDirEnv = "EVERSCRIBE_CLAUDE_SKILLS_DIR"

func newDownloadCmd() *cobra.Command {
	var dir string
	var force bool
	cmd := &cobra.Command{
		Use:   "download <skill>",
		Short: "Download a skill into your Claude Code skills directory",
		Long: "Fetches the named skill's SKILL.md from everscribe.io and writes it to\n" +
			"~/.claude/skills/<skill-dir>/SKILL.md (or --dir if provided), where\n" +
			"Claude Code reads it on next startup.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownload(cmd.Context(), cmd.OutOrStdout(), args[0], dir, force)
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "override the Claude skills root (defaults to ~/.claude/skills, or $"+DefaultSkillsDirEnv+")")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing SKILL.md without prompting")
	return cmd
}

func runDownload(ctx context.Context, stdout io.Writer, name, dirFlag string, force bool) error {
	root, err := resolveSkillsRoot(dirFlag)
	if err != nil {
		return err
	}

	f := NewFetcher()
	cat, err := f.Catalog(ctx)
	if err != nil {
		return err
	}
	skill, ok := findSkill(cat.Skills, name)
	if !ok {
		return fmt.Errorf("skill %q not found in catalog (try `es skills list`)", name)
	}

	body, err := f.SkillBody(ctx, skill.Source)
	if err != nil {
		return err
	}

	target := filepath.Join(root, skill.ClaudeCodeSkillName, "SKILL.md")
	if !force {
		if _, err := os.Stat(target); err == nil {
			return fmt.Errorf("%s already exists — re-run with --force to overwrite", target)
		}
	}
	if err := os.MkdirAll(filepath.Dir(target), skillDirMode); err != nil {
		return fmt.Errorf("create skill dir: %w", err)
	}
	if err := os.WriteFile(target, body, skillFileMode); err != nil {
		return fmt.Errorf("write SKILL.md: %w", err)
	}

	fmt.Fprintf(stdout, "Wrote %s (v%s)\n", target, skill.Version)
	fmt.Fprintf(stdout, "Restart Claude Code, then invoke /%s in your repo.\n",
		skill.ClaudeCodeSkillName)
	return nil
}

// resolveSkillsRoot picks where SKILL.md lands. Precedence:
//
//  1. --dir flag (explicit override per invocation)
//  2. $EVERSCRIBE_CLAUDE_SKILLS_DIR env (test seam / unusual configs)
//  3. $HOME/.claude/skills (Claude Code's default)
//
// Returning the directory means the caller still appends
// <claude_code_skill_name>/SKILL.md.
func resolveSkillsRoot(dirFlag string) (string, error) {
	if dirFlag != "" {
		return dirFlag, nil
	}
	if v := os.Getenv(DefaultSkillsDirEnv); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home dir: %w", err)
	}
	return filepath.Join(home, ".claude", "skills"), nil
}

// findSkill returns the catalog entry whose Name matches. Case-sensitive —
// the catalog's names are stable identifiers, not user-facing copy.
func findSkill(ss []types.Skill, name string) (types.Skill, bool) {
	for _, s := range ss {
		if s.Name == name {
			return s, true
		}
	}
	return types.Skill{}, false
}
