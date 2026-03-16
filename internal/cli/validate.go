package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nwiley/vex/internal/provider"
	"github.com/nwiley/vex/internal/spec"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <spec-file>",
		Short: "Check if a vexspec is complete",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := spec.Load(args[0])
			if err != nil {
				return err
			}

			p, err := provider.New(cfg)
			if err != nil {
				return err
			}

			result, err := spec.Validate(cmd.Context(), p, s)
			if err != nil {
				return err
			}

			out, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return fmt.Errorf("marshaling result: %w", err)
			}

			fmt.Fprintln(os.Stdout, string(out))

			if err := writeOutput("validation.json", out); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			if !result.Complete {
				os.Exit(1)
			}
			return nil
		},
	}
}
