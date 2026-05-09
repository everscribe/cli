package projects

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/cobrax"
	"github.com/everscribe/cli/internal/config"
)

func newUseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <id>",
		Short: "Set the default project so other commands can omit --project",
		Args:  cobrax.RequireArgs("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUse(cmd.Context(), cmd.OutOrStdout(), args[0])
		},
	}
	return cmd
}

func runUse(ctx context.Context, stdout io.Writer, id string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	c := client.New(pat.Token)
	// Round-trip the API once so we don't persist a typo. GetProject
	// surfaces 404 / 401 with a clear error; the user finds out before
	// every subsequent command starts hitting the wrong project.
	p, err := c.GetProject(ctx, id)
	if err != nil {
		return err
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	cfg.DefaultProjectID = p.ID
	if err := config.SaveConfig(cfg); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Default project set to %s (%s).\n", p.ID, p.Name)
	return nil
}
