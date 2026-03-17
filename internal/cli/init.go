package cli

import (
	"fmt"
	"os"

	"github.com/nickw409/vex/internal/config"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a default vex.yaml config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(".vex", 0o755); err != nil {
				return fmt.Errorf("creating .vex directory: %w", err)
			}
			if err := config.WriteDefault("vex.yaml"); err != nil {
				return err
			}
			fmt.Fprintln(cmd.ErrOrStderr(), "Created vex.yaml and .vex/")
			return nil
		},
	}
}
