package spec

import (
	"testing"
)

func TestBuildValidatePrompt(t *testing.T) {
	s := &Spec{
		Feature:     "JWT Auth",
		Description: "Token-based authentication",
		Behaviors: []Behavior{
			{Name: "login", Description: "POST /login returns JWT"},
			{Name: "refresh", Description: "POST /refresh returns new token"},
		},
	}

	prompt := buildValidatePrompt(s)

	if !contains(prompt, "JWT Auth") {
		t.Error("prompt should contain feature name")
	}
	if !contains(prompt, "Token-based authentication") {
		t.Error("prompt should contain description")
	}
	if !contains(prompt, "login") {
		t.Error("prompt should contain behavior name")
	}
	if !contains(prompt, "POST /refresh") {
		t.Error("prompt should contain behavior description")
	}
}

func TestParseValidationResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		complete bool
		count    int
	}{
		{
			name:     "complete",
			input:    `{"complete": true, "suggestions": []}`,
			complete: true,
			count:    0,
		},
		{
			name:  "incomplete",
			input: `{"complete": false, "suggestions": [{"behavior_name": "token-expiry", "description": "Add error case for expired tokens", "relation": "extends login"}]}`,
			complete: false,
			count:    1,
		},
		{
			name:     "with markdown fences",
			input:    "```json\n{\"complete\": true, \"suggestions\": []}\n```",
			complete: true,
			count:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseValidationResponse(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if result.Complete != tt.complete {
				t.Errorf("expected complete=%v, got %v", tt.complete, result.Complete)
			}
			if len(result.Suggestions) != tt.count {
				t.Errorf("expected %d suggestions, got %d", tt.count, len(result.Suggestions))
			}
		})
	}
}

func TestParseValidationResponseFields(t *testing.T) {
	input := `{"complete": false, "suggestions": [{"behavior_name": "revocation", "description": "Token revocation flow", "relation": "new"}]}`
	result, err := parseValidationResponse(input)
	if err != nil {
		t.Fatal(err)
	}
	s := result.Suggestions[0]
	if s.BehaviorName != "revocation" {
		t.Errorf("expected behavior_name 'revocation', got %q", s.BehaviorName)
	}
	if s.Relation != "new" {
		t.Errorf("expected relation 'new', got %q", s.Relation)
	}
}

func TestParseValidationResponseInvalid(t *testing.T) {
	_, err := parseValidationResponse("not json at all")
	if err == nil {
		t.Error("expected error for invalid response")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
