package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const guideText = `# Writing Vex Specifications

## When to Write a Spec
Write the vexspec BEFORE or DURING implementation, from the task description.
NEVER write a spec by reading existing code — that defeats the purpose.

## Spec Format
Create a file named <feature>.vexspec.yaml:

  feature: <Feature Name>
  description: |
    <1-2 sentence overview of what this feature does>

  behaviors:
    - name: <short-kebab-name>
      description: |
        <Natural language description of the behavior>
        <Include expected inputs, outputs, status codes>
        <Include error cases>

## Guidelines
- Each behavior should describe ONE observable external behavior
- Include error cases and edge cases (invalid input, missing data, timeouts)
- Be specific: "returns 401" not "handles errors"
- Include side effects: database writes, events emitted, files created
- Describe boundary conditions: empty lists, max lengths, concurrent access
- Do NOT describe implementation details (which function, which pattern)

## Output
Vex writes results to the .vex/ directory:
- .vex/report.json     — full check report (gaps + covered behaviors)
- .vex/validation.json — spec validation results

Always read the full report from these files. Stdout output may be
truncated by your environment. The .vex/ directory is gitignored.

## Example Workflow
1. Read task/ticket description
2. Create feature.vexspec.yaml listing all expected behaviors
3. Run: vex validate feature.vexspec.yaml
4. Review .vex/validation.json — add any missing behaviors to the spec
5. Implement code and tests
6. Run: vex check ./path/ --spec feature.vexspec.yaml
7. Read .vex/report.json for the full gap analysis
8. Fix gaps reported by vex
9. Repeat steps 6-8 until exit code 0
`

func newGuideCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "guide",
		Short: "Print instructions for writing vexspecs",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprint(cmd.OutOrStdout(), guideText)
		},
	}
}
