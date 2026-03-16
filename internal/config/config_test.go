package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Provider != "claude-cli" {
		t.Errorf("expected provider claude-cli, got %s", cfg.Provider)
	}
	if cfg.Model != "opus" {
		t.Errorf("expected model opus, got %s", cfg.Model)
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vex.yaml")

	content := []byte(`provider: claude-cli
model: haiku
languages:
  go:
    test_patterns: ["*_test.go"]
    source_patterns: ["*.go"]
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Provider != "claude-cli" {
		t.Errorf("expected provider claude-cli, got %s", cfg.Provider)
	}
	if cfg.Model != "haiku" {
		t.Errorf("expected model haiku, got %s", cfg.Model)
	}
	if lang, ok := cfg.Languages["go"]; !ok {
		t.Error("expected go language config")
	} else if len(lang.TestPatterns) != 1 || lang.TestPatterns[0] != "*_test.go" {
		t.Errorf("unexpected test patterns: %v", lang.TestPatterns)
	}
}

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vex.yaml")

	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Provider != "claude-cli" {
		t.Errorf("expected default provider claude-cli, got %s", cfg.Provider)
	}
	if cfg.Model != "opus" {
		t.Errorf("expected default model opus, got %s", cfg.Model)
	}
}

func TestLoadMissing(t *testing.T) {
	_, err := Load("/nonexistent/vex.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vex.yaml")

	if err := os.WriteFile(path, []byte(":::invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestWriteDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vex.yaml")

	if err := WriteDefault(path); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Provider != "claude-cli" {
		t.Errorf("expected provider claude-cli, got %s", cfg.Provider)
	}
}

func TestWriteDefaultAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vex.yaml")

	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := WriteDefault(path); err == nil {
		t.Error("expected error when file already exists")
	}
}

func TestWriteDefaultContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vex.yaml")

	if err := WriteDefault(path); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Provider != "claude-cli" {
		t.Errorf("expected provider claude-cli, got %s", cfg.Provider)
	}
	if cfg.Model != "opus" {
		t.Errorf("expected model opus, got %s", cfg.Model)
	}
}

func TestLoadWalksUpDirectories(t *testing.T) {
	parent := t.TempDir()
	child := filepath.Join(parent, "sub", "deep")
	if err := os.MkdirAll(child, 0755); err != nil {
		t.Fatal(err)
	}

	content := []byte("provider: claude-cli\nmodel: opus\n")
	if err := os.WriteFile(filepath.Join(parent, "vex.yaml"), content, 0644); err != nil {
		t.Fatal(err)
	}

	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(child)

	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Provider != "claude-cli" {
		t.Errorf("expected provider claude-cli, got %s", cfg.Provider)
	}
}
