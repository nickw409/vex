package provider

import (
	"testing"

	"github.com/nickw409/vex/internal/config"
)

func TestNewClaudeCLI(t *testing.T) {
	cfg := &config.Config{Provider: "claude-cli", Model: "sonnet"}
	p, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	cli, ok := p.(*ClaudeCLI)
	if !ok {
		t.Fatal("expected *ClaudeCLI")
	}
	if cli.Model != "sonnet" {
		t.Errorf("expected model sonnet, got %s", cli.Model)
	}
}

func TestNewUnknownProvider(t *testing.T) {
	cfg := &config.Config{Provider: "unknown"}
	_, err := New(cfg)
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}
