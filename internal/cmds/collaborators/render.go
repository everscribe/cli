package collaborators

import (
	"fmt"
	"io"

	"github.com/everscribe/cli/internal/output"
	"github.com/everscribe/cli/internal/types"
)

func renderCollaborator(w io.Writer, format string, col types.Collaborator) error {
	f, err := output.ParseFormat(format, true)
	if err != nil {
		return err
	}
	switch f {
	case output.FormatJSON:
		return output.JSON(w, col)
	case output.FormatYAML:
		return output.YAML(w, col)
	default:
		return collaboratorsTable(w, []types.Collaborator{col})
	}
}

func renderCollaborators(w io.Writer, format string, cols []types.Collaborator) error {
	f, err := output.ParseFormat(format, true)
	if err != nil {
		return err
	}
	switch f {
	case output.FormatJSON:
		return output.JSON(w, cols)
	case output.FormatYAML:
		return output.YAML(w, cols)
	default:
		if len(cols) == 0 {
			fmt.Fprintln(w, "No collaborators.")
			return nil
		}
		return collaboratorsTable(w, cols)
	}
}

func collaboratorsTable(w io.Writer, cols []types.Collaborator) error {
	tbl := output.NewTable(w)
	tbl.Header("EMAIL", "USERNAME", "STATUS", "PROJECTS")
	for _, col := range cols {
		tbl.Row(col.Email, col.Username, col.Status, fmt.Sprintf("%d", len(col.Projects)))
	}
	return tbl.Flush()
}
