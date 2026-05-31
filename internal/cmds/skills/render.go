package skills

import (
	"fmt"
	"io"

	"github.com/everscribe/cli/internal/output"
	"github.com/everscribe/cli/internal/types"
)

// renderSkills writes the catalog as table / JSON / YAML. JSON/YAML
// output is the bare array (no `{"skills": [...]}` envelope) — same
// convention as `es projects list` for jq/yq friendliness.
func renderSkills(w io.Writer, format string, ss []types.Skill) error {
	f, err := output.ParseFormat(format, true)
	if err != nil {
		return err
	}
	switch f {
	case output.FormatJSON:
		return output.JSON(w, ss)
	case output.FormatYAML:
		return output.YAML(w, ss)
	default:
		if len(ss) == 0 {
			fmt.Fprintln(w, "No skills.")
			return nil
		}
		return skillsTable(w, ss)
	}
}

func skillsTable(w io.Writer, ss []types.Skill) error {
	tbl := output.NewTable(w)
	tbl.Header("NAME", "VERSION", "DESCRIPTION")
	for _, s := range ss {
		tbl.Row(s.Name, s.Version, s.Description)
	}
	return tbl.Flush()
}
