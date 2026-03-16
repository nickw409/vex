package spec

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nwiley/vex/internal/provider"
)

type ValidationResult struct {
	Complete    bool                `json:"complete"`
	Suggestions []ValidationSuggestion `json:"suggestions"`
}

type ValidationSuggestion struct {
	BehaviorName string `json:"behavior_name"`
	Description  string `json:"description"`
	Relation     string `json:"relation"`
}

const validateSystemPrompt = `You are a test specification reviewer. You will receive a feature spec listing intended behaviors for a software feature. Your job is to identify behaviors that are conspicuously absent — things a user of this feature would obviously expect but that no listed behavior covers.

The spec's own feature description defines the scope boundary. Only suggest behaviors that fall within that scope.

Respond with ONLY a JSON object in this exact format:
{
  "complete": true,
  "suggestions": []
}

When suggestions are needed:
{
  "complete": false,
  "suggestions": [
    {
      "behavior_name": "suggested-kebab-name",
      "description": "What this behavior should cover",
      "relation": "new" or "extends <existing-behavior-name>"
    }
  ]
}

Rules:
- Suggest genuinely missing user-facing behaviors, not implementation details.
- Do NOT suggest: timeout handling, permission errors, graceful degradation, logging improvements, or internal error propagation paths. These are implementation concerns, not behavioral gaps.
- A behavior is "conspicuously absent" if a reasonable user of the feature would ask "wait, what happens when I do X?" and the spec has no answer.
- If a behavior covers error cases inline (e.g. "returns error when X"), that counts. Don't re-suggest it as a separate behavior.
- Prefer fewer, high-confidence suggestions over exhaustive lists.
- Use "relation": "new" for entirely missing behaviors, or "relation": "extends <name>" when an existing behavior is missing a significant aspect.`

func Validate(ctx context.Context, p provider.Provider, s *Spec) (*ValidationResult, error) {
	req := provider.CompletionRequest{
		SystemPrompt: validateSystemPrompt,
		UserPrompt:   buildValidatePrompt(s),
		MaxTokens:    4096,
	}

	resp, err := p.Complete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("validation request failed: %w", err)
	}

	return parseValidationResponse(resp.Content)
}

func buildValidatePrompt(s *Spec) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Feature: %s\n\n", s.Feature)
	if s.Description != "" {
		fmt.Fprintf(&b, "## Description\n%s\n", s.Description)
	}
	fmt.Fprintf(&b, "## Behaviors (%d defined)\n\n", len(s.Behaviors))
	for _, beh := range s.Behaviors {
		fmt.Fprintf(&b, "### %s\n%s\n", beh.Name, beh.Description)
	}
	return b.String()
}

func parseValidationResponse(content string) (*ValidationResult, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result ValidationResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("parsing validation response: %w\nraw response: %s", err, content)
	}

	return &result, nil
}
