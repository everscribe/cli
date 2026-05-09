package auth

import (
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/config"
	"github.com/everscribe/cli/internal/output"
)

func newWhoamiCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Print the user identity associated with the saved PAT",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhoami(cmd.OutOrStdout(), format)
		},
	}
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	return cmd
}

// whoamiInfo is the externally-visible shape rendered by --format=json
// or --format=yaml. Mirrors the on-disk pat.json fields minus the
// secret token.
type whoamiInfo struct {
	UserID    string    `json:"user_id" yaml:"user_id"`
	Email     string    `json:"email" yaml:"email"`
	PATID     string    `json:"pat_id" yaml:"pat_id"`
	ExpiresAt time.Time `json:"expires_at,omitempty" yaml:"expires_at,omitempty"`
}

func runWhoami(stdout io.Writer, format string) error {
	f, err := output.ParseFormat(format, true)
	if err != nil {
		return err
	}
	pat, err := config.Load()
	if err != nil {
		return err
	}

	info := whoamiInfo{
		UserID:    pat.UserID,
		Email:     pat.UserEmail,
		PATID:     pat.PATID,
		ExpiresAt: pat.ExpiresAt,
	}

	switch f {
	case output.FormatJSON:
		return output.JSON(stdout, info)
	case output.FormatYAML:
		return output.YAML(stdout, info)
	default:
		tbl := output.NewTable(stdout)
		tbl.Header("EMAIL", "USER ID", "PAT ID", "EXPIRES")
		expires := "never"
		if !pat.ExpiresAt.IsZero() {
			expires = pat.ExpiresAt.Format(time.RFC3339)
		}
		tbl.Row(pat.UserEmail, pat.UserID, pat.PATID, expires)
		return tbl.Flush()
	}
}
