package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteOutputCreatesVexDir(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	if err := writeOutput("test.json", []byte(`{"ok":true}`)); err != nil {
		t.Fatalf("writeOutput returned error: %v", err)
	}

	info, err := os.Stat(filepath.Join(dir, ".vex"))
	if err != nil {
		t.Fatalf(".vex directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal(".vex exists but is not a directory")
	}
}

func TestWriteOutputTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	content := []byte(`{"gaps":[]}`)
	if err := writeOutput("result.json", content); err != nil {
		t.Fatalf("writeOutput returned error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, ".vex", "result.json"))
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}

	if !bytes.HasSuffix(got, []byte("\n")) {
		t.Fatal("output file does not end with a trailing newline")
	}
}

func TestWriteOutputDirPermissions(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	if err := writeOutput("test.json", []byte(`{}`)); err != nil {
		t.Fatalf("writeOutput returned error: %v", err)
	}

	info, err := os.Stat(filepath.Join(dir, ".vex"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("expected .vex dir permissions 0755, got %o", info.Mode().Perm())
	}
}

func TestWriteOutputErrorWhenDirIsFile(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	// Create .vex as a regular file to block directory creation
	os.WriteFile(filepath.Join(dir, ".vex"), []byte("blocker"), 0644)

	err = writeOutput("test.json", []byte(`{}`))
	if err == nil {
		t.Error("expected error when .vex is a file, not a directory")
	}
}

func TestWriteOutputPrintsPath(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	// Capture stderr
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = origStderr })

	if err := writeOutput("out.json", []byte(`{}`)); err != nil {
		w.Close()
		t.Fatalf("writeOutput returned error: %v", err)
	}
	w.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("reading captured stderr: %v", err)
	}

	expected := filepath.Join(".vex", "out.json")
	if got := buf.String(); got != expected+"\n" {
		t.Fatalf("expected stderr %q, got %q", expected+"\n", got)
	}
}
