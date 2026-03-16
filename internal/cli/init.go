package cli

import (
	"fmt"

	"github.com/nwiley/vex/internal/config"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a default vex.yaml config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.WriteDefault("vex.yaml"); err != nil {
				return err
			}
			fmt.Fprintln(cmd.ErrOrStderr(), "Created vex.yaml")
			return nil
		},
	}
}
