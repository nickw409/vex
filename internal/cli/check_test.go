package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nickw409/vex/internal/diff"
	"github.com/nickw409/vex/internal/report"
)

func TestCheckSpecNotFound(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"check", "--spec", "/nonexistent/spec.yaml"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error for nonexistent spec file")
	}
}

func TestCheckSectionNotFound(t *testing.T) {
	dir := t.TempDir()
	vexDir := filepath.Join(dir, ".vex")
	os.MkdirAll(vexDir, 0755)

	specPath := filepath.Join(vexDir, "vexspec.yaml")
	os.WriteFile(specPath, []byte(`project: Test
sections:
  - name: Auth
    path: auth
    description: Auth module
    behaviors:
      - name: login
        description: Login endpoint
`), 0644)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"check", "--spec", specPath, "--section", "Nonexistent"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error for nonexistent section")
	}
}

func TestValidateRequiresValidSpec(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"validate", "/nonexistent/spec.yaml"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error for nonexistent spec file")
	}
}

func TestRootConfigLoadedForCheck(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	// Write a custom config
	os.WriteFile(filepath.Join(dir, "vex.yaml"), []byte("provider: claude-cli\nmodel: haiku\n"), 0644)

	// check will fail (no spec) but config should be loaded first
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"check"})
	cmd.Execute()

	if cfg == nil {
		t.Fatal("expected cfg to be set after check command")
	}
	if cfg.Model != "haiku" {
		t.Errorf("expected model 'haiku' from config, got %q", cfg.Model)
	}
}

func TestRootConfigSkippedForInit(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	// No vex.yaml — init should not try to load config
	cfg = nil
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"init"})
	cmd.Execute()

	// cfg should remain nil since init skips config loading
	if cfg != nil {
		t.Error("expected cfg to remain nil for init command")
	}
}

func TestRootConfigDefaultsOnMissing(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	// No vex.yaml — config should fall back to defaults
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"check"})
	cmd.Execute()

	if cfg == nil {
		t.Fatal("expected cfg to be set with defaults")
	}
	if cfg.Model != "opus" {
		t.Errorf("expected default model 'opus', got %q", cfg.Model)
	}
}

func TestRootConfigFlagLoadsFromPath(t *testing.T) {
	dir := t.TempDir()
	customPath := filepath.Join(dir, "custom.yaml")
	os.WriteFile(customPath, []byte("provider: claude-cli\nmodel: sonnet\n"), 0644)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"check", "--config", customPath})
	cmd.Execute()

	if cfg == nil {
		t.Fatal("expected cfg to be set")
	}
	if cfg.Model != "sonnet" {
		t.Errorf("expected model 'sonnet' from custom config, got %q", cfg.Model)
	}
}

func TestCheckNoFilesEmptyReport(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	os.WriteFile(filepath.Join(dir, "vex.yaml"), []byte("provider: claude-cli\nmodel: opus\n"), 0644)

	vexDir := filepath.Join(dir, ".vex")
	os.MkdirAll(vexDir, 0755)
	specPath := filepath.Join(vexDir, "vexspec.yaml")
	os.WriteFile(specPath, []byte(`project: Test
sections:
  - name: Empty
    path: nonexistent_dir
    description: No files here
    behaviors:
      - name: something
        description: Something
`), 0644)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"check", "--spec", specPath})

	// Should not error — produces empty report
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error for empty section, got: %v", err)
	}
}

func TestDriftSpecNotFound(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"drift", "--spec", "/nonexistent/spec.yaml"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error for nonexistent spec in drift command")
	}
}

// --- Drift command tests ---

func TestDriftNoPreviousCheck(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	// The drift command calls os.Exit(2) when no report exists.
	// We can't test os.Exit directly, but we can verify ReportModTime
	// returns zero, which is the condition that triggers exit 2.
	modTime := diff.ReportModTime(dir)
	if !modTime.IsZero() {
		t.Error("expected zero mod time when no report.json exists")
	}
}

func TestDriftAllClean(t *testing.T) {
	dir := setupGitRepoWithSpec(t)
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	// Set report.json mtime 2 seconds in the future to ensure git log --since
	// won't match the commit.
	vexDir := filepath.Join(dir, ".vex")
	os.WriteFile(filepath.Join(vexDir, "report.json"), []byte("{}"), 0644)
	future := diff.ReportModTime(dir).Add(2 * time.Second)
	os.Chtimes(filepath.Join(vexDir, "report.json"), future, future)

	// Capture stdout
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"drift", "--spec", filepath.Join(vexDir, "vexspec.yaml")})
	cmd.Execute()

	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var report struct {
		Drifted []interface{} `json:"drifted"`
		Clean   []string      `json:"clean"`
	}
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("drift output is not valid JSON: %v\noutput: %s", err, buf.String())
	}

	if len(report.Drifted) != 0 {
		t.Errorf("expected no drifted sections, got %d", len(report.Drifted))
	}
	if len(report.Clean) != 1 || report.Clean[0] != "Auth" {
		t.Errorf("expected [Auth] in clean, got %v", report.Clean)
	}
}

func setupGitRepoWithSpec(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Init git repo
	for _, args := range [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %v\n%s", args, err, out)
		}
	}

	// Create source file and commit
	srcDir := filepath.Join(dir, "internal", "auth")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "auth.go"), []byte("package auth"), 0644)

	vexDir := filepath.Join(dir, ".vex")
	os.MkdirAll(vexDir, 0755)
	os.WriteFile(filepath.Join(vexDir, "vexspec.yaml"), []byte(`project: Test
sections:
  - name: Auth
    path: internal/auth
    description: Auth module
    behaviors:
      - name: login
        description: Login endpoint
`), 0644)

	for _, args := range [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", "initial"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %v\n%s", args, err, out)
		}
	}

	return dir
}

// --- Shared behavior tests: exit-codes, error-output, json-output ---

func TestOutputReportExitZeroNoGaps(t *testing.T) {
	r := &report.Report{
		Spec:    ".vex/vexspec.yaml",
		Gaps:    []report.Gap{},
		Covered: []report.Covered{},
	}
	r.ComputeSummary(0)

	if r.HasGaps() {
		t.Error("expected HasGaps() == false for empty gaps")
	}
}

func TestOutputReportHasGapsWhenGapsExist(t *testing.T) {
	r := &report.Report{
		Spec: ".vex/vexspec.yaml",
		Gaps: []report.Gap{
			{Behavior: "login", Detail: "missing test", Suggestion: "add test"},
		},
		Covered: []report.Covered{},
	}
	r.ComputeSummary(1)

	if !r.HasGaps() {
		t.Error("expected HasGaps() == true when gaps exist")
	}
}

func TestCheckEmptyReportWritesJSON(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	os.WriteFile(filepath.Join(dir, "vex.yaml"), []byte("provider: claude-cli\nmodel: opus\n"), 0644)

	vexDir := filepath.Join(dir, ".vex")
	os.MkdirAll(vexDir, 0755)
	specPath := filepath.Join(vexDir, "vexspec.yaml")
	os.WriteFile(specPath, []byte(`project: Test
sections:
  - name: Empty
    path: nonexistent_dir
    description: No files here
    behaviors:
      - name: something
        description: Something
`), 0644)

	// Capture stdout
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"check", "--spec", specPath, "--drift=false"})
	cmd.Execute()

	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify stdout contains valid JSON
	var parsed report.Report
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\noutput: %s", err, output)
	}

	// Verify .vex/report.json was also written
	fileData, err := os.ReadFile(filepath.Join(vexDir, "report.json"))
	if err != nil {
		t.Fatalf("expected .vex/report.json to be written: %v", err)
	}
	var fileParsed report.Report
	if err := json.Unmarshal(bytes.TrimSpace(fileData), &fileParsed); err != nil {
		t.Fatalf(".vex/report.json is not valid JSON: %v", err)
	}
}

func TestCheckErrorsGoToStderr(t *testing.T) {
	// Capture stderr
	origStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Capture stdout to ensure errors don't leak there
	origStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"check", "--spec", "/nonexistent/spec.yaml", "--drift=false"})
	cmd.Execute()

	w.Close()
	wOut.Close()
	os.Stderr = origStderr
	os.Stdout = origStdout

	var stderrBuf bytes.Buffer
	stderrBuf.ReadFrom(r)

	var stdoutBuf bytes.Buffer
	stdoutBuf.ReadFrom(rOut)

	// Stdout should be empty (no JSON output on fatal error)
	if stdoutBuf.Len() > 0 {
		t.Errorf("expected no stdout on fatal error, got: %s", stdoutBuf.String())
	}
}

func TestRootConfigSkippedForGuide(t *testing.T) {
	cfg = nil
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"guide"})
	// Discard stdout
	cmd.SetOut(new(bytes.Buffer))
	cmd.Execute()

	if cfg != nil {
		t.Error("expected cfg to remain nil for guide command")
	}
}

func TestRootConfigSkippedForDrift(t *testing.T) {
	cfg = nil
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"drift", "--spec", "/nonexistent"})
	cmd.Execute()

	if cfg != nil {
		t.Error("expected cfg to remain nil for drift command")
	}
}

func TestValidateErrorGoesToStderr(t *testing.T) {
	origStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	origStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"validate", "/nonexistent/spec.yaml"})
	cmd.Execute()

	w.Close()
	wOut.Close()
	os.Stderr = origStderr
	os.Stdout = origStdout

	var stdoutBuf bytes.Buffer
	stdoutBuf.ReadFrom(rOut)

	// No JSON on stdout when spec doesn't exist
	if stdoutBuf.Len() > 0 {
		t.Errorf("expected no stdout on validate fatal error, got: %s", stdoutBuf.String())
	}
}

func TestValidateDefaultSpecPath(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	// No spec file — should error referencing default path
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"validate"})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error when default spec doesn't exist")
	}
	if !strings.Contains(err.Error(), "vexspec.yaml") {
		t.Errorf("expected error to reference vexspec.yaml, got: %v", err)
	}
}

func TestRootConfigSkippedForUpdate(t *testing.T) {
	cfg = nil
	cmd := NewRootCmd()
	// update will fail but config should not be loaded
	cmd.SetArgs([]string{"update", "--help"})
	cmd.Execute()

	if cfg != nil {
		t.Error("expected cfg to remain nil for update command")
	}
}
