package check

import (
	"testing"

	"github.com/nwiley/vex/internal/spec"
)

func TestBuildCheckPrompt(t *testing.T) {
	input := &Input{
		Spec: &spec.Spec{
			Feature:     "Auth",
			Description: "JWT authentication",
			Behaviors: []spec.Behavior{
				{Name: "login", Description: "POST /login returns JWT"},
			},
		},
		SourceFiles: map[string]string{
			"auth.go": "package auth\nfunc Login() {}",
		},
		TestFiles: map[string]string{
			"auth_test.go": "package auth\nfunc TestLogin(t *testing.T) {}",
		},
		Target:   "./auth/",
		SpecPath: "auth.vexspec.yaml",
	}

	prompt, err := buildCheckPrompt(input)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"Auth", "login", "POST /login", "auth.go", "auth_test.go"} {
		if !containsStr(prompt, want) {
			t.Errorf("prompt should contain %q", want)
		}
	}
}

func TestBuildCheckPromptTooLarge(t *testing.T) {
	large := make([]byte, maxContentSize)
	for i := range large {
		large[i] = 'x'
	}

	input := &Input{
		Spec: &spec.Spec{
			Feature:   "Test",
			Behaviors: []spec.Behavior{{Name: "b", Description: "d"}},
		},
		SourceFiles: map[string]string{"big.go": string(large)},
		TestFiles:   map[string]string{},
		Target:      ".",
		SpecPath:    "test.vexspec.yaml",
	}

	_, err := buildCheckPrompt(input)
	if err == nil {
		t.Error("expected error for oversized content")
	}
}

func TestParseCheckResponse(t *testing.T) {
	content := `{
  "gaps": [
    {"behavior": "login", "detail": "No expiry test", "suggestion": "Add TestLoginExpiry"}
  ],
  "covered": [
    {"behavior": "login", "detail": "Valid creds", "test_file": "auth_test.go", "test_name": "TestLogin"}
  ]
}`

	input := &Input{
		Spec: &spec.Spec{
			Feature: "Auth",
			Behaviors: []spec.Behavior{
				{Name: "login", Description: "login behavior"},
				{Name: "refresh", Description: "refresh behavior"},
			},
		},
		Target:   "./auth/",
		SpecPath: "auth.vexspec.yaml",
	}

	r, err := parseCheckResponse(content, input)
	if err != nil {
		t.Fatal(err)
	}

	if len(r.Gaps) != 1 {
		t.Errorf("expected 1 gap, got %d", len(r.Gaps))
	}
	if len(r.Covered) != 1 {
		t.Errorf("expected 1 covered, got %d", len(r.Covered))
	}
	if r.Summary.TotalBehaviors != 2 {
		t.Errorf("expected 2 total behaviors, got %d", r.Summary.TotalBehaviors)
	}
	if r.Target != "./auth/" {
		t.Errorf("expected target ./auth/, got %s", r.Target)
	}
}

func TestParseCheckResponseInvalid(t *testing.T) {
	input := &Input{
		Spec:     &spec.Spec{Feature: "X", Behaviors: []spec.Behavior{{Name: "a", Description: "b"}}},
		Target:   ".",
		SpecPath: "x.yaml",
	}
	_, err := parseCheckResponse("not json", input)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseCheckResponseEmpty(t *testing.T) {
	input := &Input{
		Spec:     &spec.Spec{Feature: "X", Behaviors: []spec.Behavior{{Name: "a", Description: "b"}}},
		Target:   ".",
		SpecPath: "x.yaml",
	}

	r, err := parseCheckResponse(`{"gaps": [], "covered": []}`, input)
	if err != nil {
		t.Fatal(err)
	}
	if r.HasGaps() {
		t.Error("should have no gaps")
	}
}

func TestParseCheckResponseMarkdownFenced(t *testing.T) {
	content := "```json\n{\"gaps\": [], \"covered\": [{\"behavior\": \"login\", \"detail\": \"tested\", \"test_file\": \"a.go\", \"test_name\": \"TestA\"}]}\n```"

	input := &Input{
		Spec:     &spec.Spec{Feature: "X", Behaviors: []spec.Behavior{{Name: "login", Description: "b"}}},
		Target:   ".",
		SpecPath: "x.yaml",
	}

	r, err := parseCheckResponse(content, input)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Covered) != 1 {
		t.Errorf("expected 1 covered, got %d", len(r.Covered))
	}
}

func TestParseCheckResponseNullArrays(t *testing.T) {
	content := `{"gaps": null, "covered": null}`

	input := &Input{
		Spec:     &spec.Spec{Feature: "X", Behaviors: []spec.Behavior{{Name: "a", Description: "b"}}},
		Target:   ".",
		SpecPath: "x.yaml",
	}

	r, err := parseCheckResponse(content, input)
	if err != nil {
		t.Fatal(err)
	}
	if r.Gaps == nil {
		t.Error("gaps should not be nil (should be empty slice)")
	}
	if r.Covered == nil {
		t.Error("covered should not be nil (should be empty slice)")
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
