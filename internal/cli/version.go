package cli

import (
	"fmt"

	"github.com/nwiley/vex/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	var short bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			if short {
				fmt.Println(version.Short())
			} else {
				fmt.Println(version.String())
			}
		},
	}

	cmd.Flags().BoolVarP(&short, "short", "s", false, "print version number only")

	return cmd
}
