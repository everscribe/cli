package projects

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/config"
)

func newCurrentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Print the default project ID, if one is set",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCurrent(cmd.OutOrStdout())
		},
	}
	return cmd
}

func runCurrent(stdout io.Writer) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	if cfg.DefaultProjectID == "" {
		fmt.Fprintln(stdout, "No default project set. Run `es projects use <id>` to pick one.")
		return nil
	}
	fmt.Fprintln(stdout, cfg.DefaultProjectID)
	return nil
}
