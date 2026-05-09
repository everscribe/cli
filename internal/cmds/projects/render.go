package projects

import (
	"fmt"
	"io"

	"github.com/everscribe/cli/internal/output"
	"github.com/everscribe/cli/internal/types"
)

// renderProject writes a single project. Table form is a one-row
// kubectl-style table; json/yaml emit the bare object (no list wrapper).
func renderProject(w io.Writer, format string, p types.Project) error {
	f, err := output.ParseFormat(format, true)
	if err != nil {
		return err
	}
	switch f {
	case output.FormatJSON:
		return output.JSON(w, p)
	case output.FormatYAML:
		return output.YAML(w, p)
	default:
		return projectsTable(w, []types.Project{p})
	}
}

// renderProjects writes a slice of projects. JSON/YAML output is the
// bare array (no `{"projects": [...]}` envelope) — friendlier for jq
// and yq scripting.
func renderProjects(w io.Writer, format string, ps []types.Project) error {
	f, err := output.ParseFormat(format, true)
	if err != nil {
		return err
	}
	switch f {
	case output.FormatJSON:
		return output.JSON(w, ps)
	case output.FormatYAML:
		return output.YAML(w, ps)
	default:
		if len(ps) == 0 {
			fmt.Fprintln(w, "No projects.")
			return nil
		}
		return projectsTable(w, ps)
	}
}

func projectsTable(w io.Writer, ps []types.Project) error {
	tbl := output.NewTable(w)
	tbl.Header("ID", "NAME", "AGE")
	for _, p := range ps {
		tbl.Row(p.ID, p.Name, output.Age(p.CreatedAt))
	}
	return tbl.Flush()
}
