package spec

import (
	"os"
	"path/filepath"
	"testing"
)

func writeSpec(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.vexspec.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad(t *testing.T) {
	path := writeSpec(t, `feature: JWT Auth
description: Token-based authentication
behaviors:
  - name: login
    description: POST /login returns JWT on valid credentials
  - name: token-validation
    description: Protected endpoints require valid JWT
`)

	s, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if s.Feature != "JWT Auth" {
		t.Errorf("expected feature 'JWT Auth', got %q", s.Feature)
	}
	if len(s.Behaviors) != 2 {
		t.Errorf("expected 2 behaviors, got %d", len(s.Behaviors))
	}
	if s.Behaviors[0].Name != "login" {
		t.Errorf("expected first behavior 'login', got %q", s.Behaviors[0].Name)
	}
}

func TestLoadMissingFeature(t *testing.T) {
	path := writeSpec(t, `behaviors:
  - name: login
    description: does something
`)

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for missing feature")
	}
}

func TestLoadEmptyBehaviors(t *testing.T) {
	path := writeSpec(t, `feature: Test
behaviors: []
`)

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for empty behaviors")
	}
}

func TestLoadBehaviorMissingName(t *testing.T) {
	path := writeSpec(t, `feature: Test
behaviors:
  - description: does something
`)

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for behavior missing name")
	}
}

func TestLoadBehaviorMissingDescription(t *testing.T) {
	path := writeSpec(t, `feature: Test
behaviors:
  - name: login
`)

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for behavior missing description")
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	_, err := Load("/nonexistent/spec.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	path := writeSpec(t, ":::invalid yaml\n\t{{{")

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}
