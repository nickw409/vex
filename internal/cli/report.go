package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nickw409/vex/internal/report"
	"github.com/spf13/cobra"
)

func newReportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Display a formatted summary of the last check",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := filepath.Join(vexDir, "report.json")
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("reading report: %w (run vex check first)", err)
			}

			var r report.Report
			if err := json.Unmarshal(data, &r); err != nil {
				return fmt.Errorf("parsing report: %w", err)
			}

			printReport(cmd, &r)
			return nil
		},
	}

	return cmd
}

func printReport(cmd *cobra.Command, r *report.Report) {
	w := cmd.OutOrStdout()

	// Summary
	fmt.Fprintf(w, "%d behaviors: %d covered, %d gaps\n\n",
		r.Summary.TotalBehaviors, r.Summary.FullyCovered, r.Summary.GapsFound)

	if len(r.Gaps) == 0 {
		fmt.Fprintln(w, "No gaps found.")
		return
	}

	// Group gaps by behavior
	type group struct {
		details []report.Gap
	}
	groups := make(map[string]*group)
	var order []string

	for _, g := range r.Gaps {
		gr, ok := groups[g.Behavior]
		if !ok {
			gr = &group{}
			groups[g.Behavior] = gr
			order = append(order, g.Behavior)
		}
		gr.details = append(gr.details, g)
	}
	sort.Strings(order)

	for _, name := range order {
		gr := groups[name]
		fmt.Fprintf(w, "%s (%d)\n", name, len(gr.details))
		for _, g := range gr.details {
			detail := g.Detail
			if len(detail) > 120 {
				detail = detail[:117] + "..."
			}
			detail = strings.ReplaceAll(detail, "\n", " ")
			fmt.Fprintf(w, "  - %s\n", detail)
		}
		fmt.Fprintln(w)
	}
}
