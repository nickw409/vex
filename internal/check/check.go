package check

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nwiley/vex/internal/provider"
	"github.com/nwiley/vex/internal/report"
	"github.com/nwiley/vex/internal/spec"
)

const maxContentSize = 400_000

type Input struct {
	Spec        *spec.Spec
	SourceFiles map[string]string
	TestFiles   map[string]string
	Target      string
	SpecPath    string
}

const checkSystemPrompt = `You are a test coverage auditor. You will receive:
1. A feature specification with named behaviors describing intended functionality
2. Source code files implementing the feature
3. Test files testing the feature

Your job: determine whether the tests adequately cover each behavior described in the specification.

Respond with ONLY a JSON object in this exact format:
{
  "gaps": [
    {
      "behavior": "behavior-name",
      "detail": "What specific aspect is not tested",
      "suggestion": "What test to add"
    }
  ],
  "covered": [
    {
      "behavior": "behavior-name",
      "detail": "What aspect is covered",
      "test_file": "filename_test.go",
      "test_name": "TestFunctionName"
    }
  ]
}

Rules:
- A behavior can have MULTIPLE covered entries AND multiple gap entries (partial coverage).
- Only flag genuine gaps. If the behavior is tested, mark it covered.
- Be specific about which test file and test function covers each behavior.
- For gaps, suggest concrete test names and what they should assert.
- Do NOT invent behaviors beyond what the spec describes. The spec is the scope boundary.
- If a behavior has sub-points, each sub-point can be a separate covered/gap entry, but the behavior name stays the same.`

type checkResponse struct {
	Gaps    []report.Gap     `json:"gaps"`
	Covered []report.Covered `json:"covered"`
}

func Run(ctx context.Context, p provider.Provider, input *Input) (*report.Report, error) {
	userPrompt, err := buildCheckPrompt(input)
	if err != nil {
		return nil, err
	}

	req := provider.CompletionRequest{
		SystemPrompt: checkSystemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    8192,
	}

	resp, err := p.Complete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("check request failed: %w", err)
	}

	return parseCheckResponse(resp.Content, input)
}

func buildCheckPrompt(input *Input) (string, error) {
	var b strings.Builder

	fmt.Fprintf(&b, "## Specification\n\nFeature: %s\n", input.Spec.Feature)
	if input.Spec.Description != "" {
		fmt.Fprintf(&b, "%s\n", input.Spec.Description)
	}
	b.WriteString("\n### Behaviors\n\n")
	for _, beh := range input.Spec.Behaviors {
		fmt.Fprintf(&b, "#### %s\n%s\n\n", beh.Name, beh.Description)
	}

	b.WriteString("## Source Files\n\n")
	for name, content := range input.SourceFiles {
		fmt.Fprintf(&b, "### %s\n```\n%s\n```\n\n", name, content)
	}

	b.WriteString("## Test Files\n\n")
	for name, content := range input.TestFiles {
		fmt.Fprintf(&b, "### %s\n```\n%s\n```\n\n", name, content)
	}

	result := b.String()
	if len(result) > maxContentSize {
		return "", fmt.Errorf("total file content exceeds %d chars; use --diff or a narrower target path", maxContentSize)
	}

	return result, nil
}

func parseCheckResponse(content string, input *Input) (*report.Report, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var resp checkResponse
	if err := json.Unmarshal([]byte(content), &resp); err != nil {
		return nil, fmt.Errorf("parsing check response: %w\nraw response: %s", err, content)
	}

	r := &report.Report{
		Target:  input.Target,
		Spec:    input.SpecPath,
		Gaps:    resp.Gaps,
		Covered: resp.Covered,
	}

	if r.Gaps == nil {
		r.Gaps = []report.Gap{}
	}
	if r.Covered == nil {
		r.Covered = []report.Covered{}
	}

	r.ComputeSummary(len(input.Spec.Behaviors))

	return r, nil
}
