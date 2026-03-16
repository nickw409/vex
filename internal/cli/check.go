package cli

import (
	"fmt"
	"os"

	"github.com/nwiley/vex/internal/check"
	"github.com/nwiley/vex/internal/diff"
	"github.com/nwiley/vex/internal/lang"
	"github.com/nwiley/vex/internal/provider"
	"github.com/nwiley/vex/internal/report"
	"github.com/nwiley/vex/internal/spec"
	"github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
	var specPath string
	var useDiff bool

	cmd := &cobra.Command{
		Use:   "check [target]",
		Short: "Check test coverage against a vexspec",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if specPath == "" {
				return fmt.Errorf("--spec is required")
			}

			target := "."
			if len(args) > 0 {
				target = args[0]
			} else if !useDiff {
				return fmt.Errorf("target directory is required (or use --diff)")
			}

			s, err := spec.Load(specPath)
			if err != nil {
				return err
			}

			l, err := lang.Detect(target, cfg.Languages)
			if err != nil && !useDiff {
				return err
			}

			var sourceFiles, testFiles []string

			if useDiff {
				changed, err := diff.ChangedFiles(target)
				if err != nil {
					return err
				}

				if l == nil {
					// No language detected and no files changed — empty report
					return emptyReport(s, target, specPath)
				}

				sourceFiles, testFiles = diff.FilterByLanguage(changed, l)

				if len(sourceFiles) == 0 && len(testFiles) == 0 {
					return emptyReport(s, target, specPath)
				}
			} else {
				sourceFiles, testFiles, err = lang.FindFiles(target, l)
				if err != nil {
					return err
				}

				if len(testFiles) == 0 {
					return fmt.Errorf("no test files found in %s", target)
				}
			}

			srcMap, err := readFiles(sourceFiles)
			if err != nil {
				return err
			}
			testMap, err := readFiles(testFiles)
			if err != nil {
				return err
			}

			p, err := provider.New(cfg)
			if err != nil {
				return err
			}

			input := &check.Input{
				Spec:        s,
				SourceFiles: srcMap,
				TestFiles:   testMap,
				Target:      target,
				SpecPath:    specPath,
			}

			r, err := check.Run(cmd.Context(), p, input)
			if err != nil {
				return err
			}

			return outputReport(r)
		},
	}

	cmd.Flags().StringVar(&specPath, "spec", "", "path to vexspec.yaml (required)")
	cmd.Flags().BoolVar(&useDiff, "diff", false, "scope check to git diff only")

	return cmd
}

func readFiles(paths []string) (map[string]string, error) {
	m := make(map[string]string, len(paths))
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", p, err)
		}
		m[p] = string(data)
	}
	return m, nil
}

func emptyReport(s *spec.Spec, target, specPath string) error {
	r := &report.Report{
		Target:  target,
		Spec:    specPath,
		Gaps:    []report.Gap{},
		Covered: []report.Covered{},
	}
	r.ComputeSummary(len(s.Behaviors))
	return outputReport(r)
}

func outputReport(r *report.Report) error {
	out, err := r.JSON()
	if err != nil {
		return fmt.Errorf("marshaling report: %w", err)
	}

	fmt.Fprintln(os.Stdout, string(out))

	if err := writeOutput("report.json", out); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	if r.HasGaps() {
		os.Exit(1)
	}
	return nil
}
