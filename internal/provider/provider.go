package provider

import (
	"context"
	"fmt"

	"github.com/nwiley/vex/internal/config"
)

type CompletionRequest struct {
	SystemPrompt string
	UserPrompt   string
	MaxTokens    int
}

type CompletionResponse struct {
	Content string
	Usage   TokenUsage
}

type TokenUsage struct {
	InputTokens  int
	OutputTokens int
	CostUSD      float64
	DurationMS   int
}

type Provider interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
}

func New(cfg *config.Config) (Provider, error) {
	switch cfg.Provider {
	case "claude-cli":
		return &ClaudeCLI{Model: cfg.Model}, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider)
	}
}
