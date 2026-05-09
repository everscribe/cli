package keys

import (
	"errors"

	"github.com/spf13/cobra"
)

func newRevokeCmd() *cobra.Command {
	var (
		project string
		key     string
	)
	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke an ingest API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = project
			_ = key
			return errors.New("not implemented (step 5)")
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (required)")
	cmd.Flags().StringVar(&key, "key", "", "key ID (required)")
	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}
