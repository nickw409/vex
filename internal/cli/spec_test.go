package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nickw409/vex/internal/provider"
	"github.com/nickw409/vex/internal/spec"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type mockProvider struct {
	response string
	err      error
}

func (m *mockProvider) Complete(ctx context.Context, req provider.CompletionRequest) (provider.CompletionResponse, error) {
	if m.err != nil {
		return provider.CompletionResponse{}, m.err
	}
	return provider.CompletionResponse{Content: m.response}, nil
}

func TestSpecRequiresDescription(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"spec"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no description argument provided")
	}
}

func TestSpecGenerateNewCreatesSpec(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	sections := []spec.Section{
		{
			Name:        "Auth",
			Description: "Authentication module",
			Behaviors: []spec.Behavior{
				{Name: "login", Description: "Authenticates user"},
				{Name: "logout", Description: "Logs user out"},
			},
		},
		{
			Name:        "Users",
			Description: "User management",
			Behaviors: []spec.Behavior{
				{Name: "create-user", Description: "Creates a new user"},
			},
		},
	}

	specPath := ".vex/vexspec.yaml"
	if err := createSpec(specPath, sections); err != nil {
		t.Fatalf("createSpec failed: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Fatal("expected .vex/vexspec.yaml to be created")
	}

	// Load and verify content
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("reading spec: %v", err)
	}

	var ps spec.ProjectSpec
	if err := yaml.Unmarshal(data, &ps); err != nil {
		t.Fatalf("parsing spec: %v", err)
	}

	// Project name should be inferred from directory basename
	expectedProject := filepath.Base(dir)
	if ps.Project != expectedProject {
		t.Errorf("expected project name %q, got %q", expectedProject, ps.Project)
	}

	if len(ps.Sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(ps.Sections))
	}
	if ps.Sections[0].Name != "Auth" {
		t.Errorf("expected first section name 'Auth', got %q", ps.Sections[0].Name)
	}
	if ps.Sections[1].Name != "Users" {
		t.Errorf("expected second section name 'Users', got %q", ps.Sections[1].Name)
	}
}

func TestSpecGenerateAppendPreservesExisting(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	// Create an initial spec
	os.MkdirAll(".vex", 0755)
	specPath := ".vex/vexspec.yaml"
	initialSpec := spec.ProjectSpec{
		Project: "TestProject",
		Sections: []spec.Section{
			{
				Name:        "Auth",
				Description: "Authentication module",
				Behaviors: []spec.Behavior{
					{Name: "login", Description: "Authenticates user"},
				},
			},
		},
	}
	initialData, _ := yaml.Marshal(&initialSpec)
	os.WriteFile(specPath, initialData, 0644)

	// Load existing and append new sections
	existing, err := spec.LoadProject(specPath)
	if err != nil {
		t.Fatalf("loading existing spec: %v", err)
	}

	newSections := []spec.Section{
		{
			Name:        "Payments",
			Description: "Payment processing",
			Behaviors: []spec.Behavior{
				{Name: "charge", Description: "Charges a card"},
			},
		},
	}

	if err := appendSections(specPath, existing, newSections); err != nil {
		t.Fatalf("appendSections failed: %v", err)
	}

	// Reload and verify
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("reading spec: %v", err)
	}

	var ps spec.ProjectSpec
	if err := yaml.Unmarshal(data, &ps); err != nil {
		t.Fatalf("parsing spec: %v", err)
	}

	if len(ps.Sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(ps.Sections))
	}

	// Original section preserved
	if ps.Sections[0].Name != "Auth" {
		t.Errorf("expected first section 'Auth', got %q", ps.Sections[0].Name)
	}
	if len(ps.Sections[0].Behaviors) != 1 || ps.Sections[0].Behaviors[0].Name != "login" {
		t.Error("original Auth section behaviors not preserved")
	}

	// New section added
	if ps.Sections[1].Name != "Payments" {
		t.Errorf("expected second section 'Payments', got %q", ps.Sections[1].Name)
	}
}

func TestSpecExtendAddsBehaviors(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	// Create an initial spec with one section
	os.MkdirAll(".vex", 0755)
	specPath := ".vex/vexspec.yaml"
	initialSpec := spec.ProjectSpec{
		Project: "TestProject",
		Sections: []spec.Section{
			{
				Name:        "Auth",
				Description: "Authentication module",
				Behaviors: []spec.Behavior{
					{Name: "login", Description: "Authenticates user"},
				},
			},
		},
	}
	initialData, _ := yaml.Marshal(&initialSpec)
	os.WriteFile(specPath, initialData, 0644)

	// Mock provider returns YAML with new behaviors
	mock := &mockProvider{
		response: `behaviors:
  - name: password-reset
    description: Sends password reset email
  - name: token-refresh
    description: Refreshes expired auth tokens`,
	}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	err := extendSection(cmd, mock, specPath, "Auth", "add password reset and token refresh")
	if err != nil {
		t.Fatalf("extendSection failed: %v", err)
	}

	// Reload and verify
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("reading spec: %v", err)
	}

	var ps spec.ProjectSpec
	if err := yaml.Unmarshal(data, &ps); err != nil {
		t.Fatalf("parsing spec: %v", err)
	}

	if len(ps.Sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(ps.Sections))
	}

	behaviors := ps.Sections[0].Behaviors
	if len(behaviors) != 3 {
		t.Fatalf("expected 3 behaviors (1 original + 2 new), got %d", len(behaviors))
	}

	if behaviors[0].Name != "login" {
		t.Errorf("expected original behavior 'login', got %q", behaviors[0].Name)
	}
	if behaviors[1].Name != "password-reset" {
		t.Errorf("expected new behavior 'password-reset', got %q", behaviors[1].Name)
	}
	if behaviors[2].Name != "token-refresh" {
		t.Errorf("expected new behavior 'token-refresh', got %q", behaviors[2].Name)
	}
}

func TestSpecExtendSectionNotFound(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	// Create a spec with one section
	os.MkdirAll(".vex", 0755)
	specPath := ".vex/vexspec.yaml"
	initialSpec := spec.ProjectSpec{
		Project: "TestProject",
		Sections: []spec.Section{
			{
				Name:        "Auth",
				Description: "Authentication module",
				Behaviors: []spec.Behavior{
					{Name: "login", Description: "Authenticates user"},
				},
			},
		},
	}
	initialData, _ := yaml.Marshal(&initialSpec)
	os.WriteFile(specPath, initialData, 0644)

	mock := &mockProvider{response: ""}
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	err := extendSection(cmd, mock, specPath, "NonExistent", "some description")
	if err == nil {
		t.Fatal("expected error when extending nonexistent section")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error message, got %q", err.Error())
	}
}

func TestSpecExtendPrintsCount(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	// Create a spec
	os.MkdirAll(".vex", 0755)
	specPath := ".vex/vexspec.yaml"
	initialSpec := spec.ProjectSpec{
		Project: "TestProject",
		Sections: []spec.Section{
			{
				Name:        "Auth",
				Description: "Authentication module",
				Behaviors: []spec.Behavior{
					{Name: "login", Description: "Authenticates user"},
				},
			},
		},
	}
	initialData, _ := yaml.Marshal(&initialSpec)
	os.WriteFile(specPath, initialData, 0644)

	mock := &mockProvider{
		response: `behaviors:
  - name: password-reset
    description: Sends password reset email
  - name: token-refresh
    description: Refreshes expired auth tokens
  - name: mfa-verify
    description: Verifies MFA code`,
	}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	err := extendSection(cmd, mock, specPath, "Auth", "add auth features")
	w.Close()
	os.Stderr = oldStderr

	if err != nil {
		t.Fatalf("extendSection failed: %v", err)
	}

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "3 behavior(s)") {
		t.Errorf("expected stderr to mention '3 behavior(s)', got %q", output)
	}
}

func TestInferProjectName(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	name := inferProjectName()
	expected := filepath.Base(dir)
	if name != expected {
		t.Errorf("expected project name %q, got %q", expected, name)
	}
}
