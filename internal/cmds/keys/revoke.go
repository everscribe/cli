package keys

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/config"
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
			return runRevoke(cmd.Context(), cmd.OutOrStdout(), project, key)
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (required)")
	cmd.Flags().StringVar(&key, "key", "", "key ID (required)")
	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}

func runRevoke(ctx context.Context, stdout io.Writer, projectID, keyID string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	c := client.New(pat.Token)
	if err := c.RevokeAPIKey(ctx, projectID, keyID); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Key %s revoked.\n", keyID)
	return nil
}
