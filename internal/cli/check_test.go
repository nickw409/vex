package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckRequiresSpecFlag(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"check", "."})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error when --spec is missing")
	}
}

func TestCheckRequiresTarget(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"check", "--spec", "some.yaml"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error when target is missing and --diff not set")
	}
}

func TestCheckSpecFileNotFound(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"check", ".", "--spec", "/nonexistent/spec.yaml"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error for nonexistent spec file")
	}
}

func TestCheckNoTestFilesFound(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "test.vexspec.yaml")
	os.WriteFile(specPath, []byte(`feature: Test
behaviors:
  - name: foo
    description: does something
`), 0644)

	// Create a source file but no test files
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"check", dir, "--spec", specPath})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error when no test files found")
	}
}

func TestCheckValidateRequiresArg(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"validate"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error when no spec file arg provided")
	}
}

func TestCheckValidateSpecNotFound(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"validate", "/nonexistent/spec.yaml"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error for nonexistent spec file")
	}
}
