package types

// Skill is one entry in the Everscribe skills catalog
// (everscribe.io/manifests/skills.json). The fields mirror the JSON
// served by the monorepo's internal/ui/manifests/ package — keep
// the two in sync when the catalog schema changes.
type Skill struct {
	// Name is the catalog identifier (`setup`). Stable across versions —
	// callers refer to a skill by this name in `es skills download <name>`.
	Name string `json:"name"`

	// Description is a short one-liner shown in `es skills list`.
	Description string `json:"description"`

	// Version is the skill body's content version. Bumped when the
	// SKILL.md is materially updated. Displayed in `list`.
	Version string `json:"version"`

	// Source is the URL of the actual SKILL.md content. `download`
	// fetches this URL directly and writes the body to the user's
	// Claude Code skills directory.
	Source string `json:"source"`

	// ClaudeCodeSkillName is the directory name Claude Code recognizes
	// under ~/.claude/skills/<name>/. For the setup skill this is
	// `everscribe-setup`; `download` writes to
	// ~/.claude/skills/everscribe-setup/SKILL.md so Claude Code picks
	// it up automatically.
	ClaudeCodeSkillName string `json:"claude_code_skill_name"`
}

// SkillCatalog is the top-level shape of skills.json — a versioned
// schema plus the array of skills. The schema number lets us version
// the catalog format independently of individual skill versions.
type SkillCatalog struct {
	Schema int     `json:"schema"`
	Skills []Skill `json:"skills"`
}
