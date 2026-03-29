package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSpec(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "vexspec.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadProject(t *testing.T) {
	path := writeSpec(t, `project: MyApp
description: Test project
sections:
  - name: Auth
    path: internal/auth
    description: Authentication module
    behaviors:
      - name: login
        description: POST /login returns JWT
      - name: logout
        description: POST /logout invalidates session
`)

	ps, err := LoadProject(path)
	if err != nil {
		t.Fatal(err)
	}

	if ps.Project != "MyApp" {
		t.Errorf("expected project 'MyApp', got %q", ps.Project)
	}
	if len(ps.Sections) != 1 {
		t.Errorf("expected 1 section, got %d", len(ps.Sections))
	}
	if ps.Sections[0].Name != "Auth" {
		t.Errorf("expected section name 'Auth', got %q", ps.Sections[0].Name)
	}
	if len(ps.Sections[0].Behaviors) != 2 {
		t.Errorf("expected 2 behaviors, got %d", len(ps.Sections[0].Behaviors))
	}
}

func TestLoadProjectWithShared(t *testing.T) {
	path := writeSpec(t, `project: MyApp
shared:
  - name: error-handling
    description: All errors return structured JSON
sections:
  - name: Auth
    path: internal/auth
    description: Auth module
    shared: [error-handling]
    behaviors:
      - name: login
        description: Login endpoint
`)

	ps, err := LoadProject(path)
	if err != nil {
		t.Fatal(err)
	}

	all := ps.AllBehaviors(&ps.Sections[0])
	if len(all) != 2 {
		t.Errorf("expected 2 behaviors (1 shared + 1 own), got %d", len(all))
	}
	if all[0].Name != "error-handling" {
		t.Errorf("expected shared behavior first, got %q", all[0].Name)
	}
}

func TestLoadProjectWithSubsections(t *testing.T) {
	path := writeSpec(t, `project: MyApp
sections:
  - name: App Server
    path: app
    description: Main server
    behaviors:
      - name: websocket
        description: WebSocket handling
    subsections:
      - name: Auth Handlers
        file: app/handlers/auth.go
        behaviors:
          - name: login
            description: Login handler
      - name: Analysis
        path: [app/handlers/analysis/, app/db/queries/]
        behaviors:
          - name: result-queries
            description: Query results
`)

	ps, err := LoadProject(path)
	if err != nil {
		t.Fatal(err)
	}

	sec := &ps.Sections[0]
	all := ps.AllBehaviors(sec)
	if len(all) != 3 {
		t.Errorf("expected 3 behaviors (1 section + 2 subsection), got %d", len(all))
	}

	paths := SectionPaths(sec)
	// app (section path), app/handlers/analysis/, app/db/queries/ (subsection paths)
	if len(paths) != 3 {
		t.Errorf("expected 3 paths, got %d: %v", len(paths), paths)
	}

	files := SectionFiles(sec)
	// app/handlers/auth.go (subsection file)
	if len(files) != 1 || files[0] != "app/handlers/auth.go" {
		t.Errorf("expected [app/handlers/auth.go], got %v", files)
	}

	allPaths := SectionAllPaths(sec)
	// app, app/handlers (from file dir), app/handlers/analysis/, app/db/queries/
	if len(allPaths) != 4 {
		t.Errorf("expected 4 allPaths, got %d: %v", len(allPaths), allPaths)
	}
}

func TestLoadProjectPathAsString(t *testing.T) {
	path := writeSpec(t, `project: MyApp
sections:
  - name: Core
    path: internal/core
    description: Core module
    behaviors:
      - name: process
        description: Process data
`)

	ps, err := LoadProject(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(ps.Sections[0].Path) != 1 || ps.Sections[0].Path[0] != "internal/core" {
		t.Errorf("expected path ['internal/core'], got %v", ps.Sections[0].Path)
	}
}

func TestLoadProjectPathAsList(t *testing.T) {
	path := writeSpec(t, `project: MyApp
sections:
  - name: Auth
    path: [src/handlers/auth.go, src/auth/]
    description: Auth module
    behaviors:
      - name: login
        description: Login
`)

	ps, err := LoadProject(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(ps.Sections[0].Path) != 2 {
		t.Errorf("expected 2 paths, got %d", len(ps.Sections[0].Path))
	}
}

func TestLoadProjectMissingProject(t *testing.T) {
	path := writeSpec(t, `sections:
  - name: Auth
    description: Auth
    behaviors:
      - name: login
        description: Login
`)

	_, err := LoadProject(path)
	if err == nil {
		t.Error("expected error for missing project name")
	}
}

func TestLoadProjectNoSections(t *testing.T) {
	path := writeSpec(t, `project: MyApp
sections: []
`)

	_, err := LoadProject(path)
	if err == nil {
		t.Error("expected error for empty sections")
	}
}

func TestLoadProjectUnknownSharedRef(t *testing.T) {
	path := writeSpec(t, `project: MyApp
sections:
  - name: Auth
    description: Auth
    shared: [nonexistent]
    behaviors:
      - name: login
        description: Login
`)

	_, err := LoadProject(path)
	if err == nil {
		t.Error("expected error for unknown shared reference")
	}
}

func TestLoadProjectSubsectionPathAndFile(t *testing.T) {
	path := writeSpec(t, `project: MyApp
sections:
  - name: Auth
    description: Auth
    behaviors:
      - name: login
        description: Login
    subsections:
      - name: Bad
        path: some/dir
        file: some/file.go
        behaviors:
          - name: thing
            description: Thing
`)

	_, err := LoadProject(path)
	if err == nil {
		t.Error("expected error when subsection has both path and file")
	}
}

func TestLoadProjectBehaviorMissingName(t *testing.T) {
	path := writeSpec(t, `project: MyApp
sections:
  - name: Auth
    description: Auth
    behaviors:
      - description: Login
`)

	_, err := LoadProject(path)
	if err == nil {
		t.Error("expected error for behavior missing name")
	}
}

func TestLoadProjectBehaviorMissingDescription(t *testing.T) {
	path := writeSpec(t, `project: MyApp
sections:
  - name: Auth
    description: Auth
    behaviors:
      - name: login
`)

	_, err := LoadProject(path)
	if err == nil {
		t.Error("expected error for behavior missing description")
	}
}

func TestLoadProjectInvalidYAML(t *testing.T) {
	path := writeSpec(t, ":::invalid yaml\n\t{{{")

	_, err := LoadProject(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadProjectNonexistentFile(t *testing.T) {
	_, err := LoadProject("/nonexistent/spec.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadProjectDefaultPath(t *testing.T) {
	// LoadProject("") should look for .vex/vexspec.yaml
	_, err := LoadProject("")
	if err == nil {
		t.Error("expected error since .vex/vexspec.yaml doesn't exist in test dir")
	}
}

func TestLoadProjectDuplicateSharedBehaviorNames(t *testing.T) {
	path := writeSpec(t, `project: MyApp
shared:
  - name: error-handling
    description: Structured errors
  - name: error-handling
    description: Duplicate shared behavior
sections:
  - name: Auth
    description: Auth
    behaviors:
      - name: login
        description: Login
`)

	_, err := LoadProject(path)
	if err == nil {
		t.Error("expected error for duplicate shared behavior names")
	}
	if err != nil && !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("expected error to mention 'duplicate', got: %s", err.Error())
	}
}

func TestLoadProjectSharedBehaviorMissingDescription(t *testing.T) {
	path := writeSpec(t, `project: MyApp
shared:
  - name: error-handling
sections:
  - name: Auth
    description: Auth
    behaviors:
      - name: login
        description: Login
`)

	_, err := LoadProject(path)
	if err == nil {
		t.Error("expected error for shared behavior missing description")
	}
}

func TestLoadProjectSharedBehaviorMissingName(t *testing.T) {
	path := writeSpec(t, `project: MyApp
shared:
  - description: Structured errors
sections:
  - name: Auth
    description: Auth
    behaviors:
      - name: login
        description: Login
`)

	_, err := LoadProject(path)
	if err == nil {
		t.Error("expected error for shared behavior missing name")
	}
}

func TestLoadProjectSubsectionMissingName(t *testing.T) {
	path := writeSpec(t, `project: MyApp
sections:
  - name: Auth
    description: Auth
    behaviors:
      - name: login
        description: Login
    subsections:
      - path: some/dir
        behaviors:
          - name: thing
            description: Thing
`)

	_, err := LoadProject(path)
	if err == nil {
		t.Error("expected error for subsection missing name")
	}
}

func TestLoadProjectSectionMissingName(t *testing.T) {
	path := writeSpec(t, `project: MyApp
sections:
  - description: Auth module
    behaviors:
      - name: login
        description: Login
`)

	_, err := LoadProject(path)
	if err == nil {
		t.Error("expected error for section missing name")
	}
}

func TestResolveShared(t *testing.T) {
	ps := &ProjectSpec{
		Project: "Test",
		Shared: []Behavior{
			{Name: "err", Description: "error handling"},
			{Name: "log", Description: "logging"},
		},
		Sections: []Section{
			{Name: "A", Shared: []string{"err"}},
		},
	}

	resolved := ps.ResolveShared(&ps.Sections[0])
	if len(resolved) != 1 {
		t.Fatalf("expected 1 resolved shared behavior, got %d", len(resolved))
	}
	if resolved[0].Name != "err" {
		t.Errorf("expected resolved[0].Name=err, got %s", resolved[0].Name)
	}
}

func TestSectionChecksum_Stable(t *testing.T) {
	sec := &Section{
		Name:        "Auth",
		Description: "auth module",
		Behaviors:   []Behavior{{Name: "login", Description: "POST /login"}},
	}
	sum1 := SectionChecksum(sec, nil)
	sum2 := SectionChecksum(sec, nil)
	if sum1 != sum2 {
		t.Errorf("checksum not stable: %s != %s", sum1, sum2)
	}
}

func TestSectionChecksum_ChangesOnEdit(t *testing.T) {
	sec := &Section{
		Name:        "Auth",
		Description: "auth module",
		Behaviors:   []Behavior{{Name: "login", Description: "POST /login"}},
	}
	sum1 := SectionChecksum(sec, nil)

	sec.Behaviors[0].Description = "POST /login returns JWT"
	sum2 := SectionChecksum(sec, nil)
	if sum1 == sum2 {
		t.Error("checksum should change when behavior description changes")
	}
}

func TestSectionChecksum_IncludesShared(t *testing.T) {
	sec := &Section{
		Name:        "Auth",
		Description: "auth module",
		Shared:      []string{"err"},
	}
	shared1 := []Behavior{{Name: "err", Description: "returns 500"}}
	shared2 := []Behavior{{Name: "err", Description: "returns 500 with trace"}}

	sum1 := SectionChecksum(sec, shared1)
	sum2 := SectionChecksum(sec, shared2)
	if sum1 == sum2 {
		t.Error("checksum should change when shared behavior description changes")
	}
}

func TestSectionChecksum_IncludesSubsections(t *testing.T) {
	sec := &Section{
		Name:        "Auth",
		Description: "auth module",
		Subsections: []Subsection{
			{Name: "Token", Behaviors: []Behavior{{Name: "refresh", Description: "refresh token"}}},
		},
	}
	sum1 := SectionChecksum(sec, nil)

	sec.Subsections[0].Behaviors[0].Description = "refresh access token"
	sum2 := SectionChecksum(sec, nil)
	if sum1 == sum2 {
		t.Error("checksum should change when subsection behavior changes")
	}
}

func TestLoadProjectWithCoveredOverrides(t *testing.T) {
	path := writeSpec(t, `project: MyApp
sections:
  - name: Worker
    path: src/worker
    description: Worker process
    covered:
      - behavior: serve-loop
        reason: tested via e2e binary spawn in tests/e2e/worker_test.rs
    behaviors:
      - name: serve-loop
        description: Main event loop
      - name: shutdown
        description: Graceful shutdown
`)

	ps, err := LoadProject(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(ps.Sections[0].Covered) != 1 {
		t.Fatalf("expected 1 covered override, got %d", len(ps.Sections[0].Covered))
	}
	co := ps.Sections[0].Covered[0]
	if co.Behavior != "serve-loop" {
		t.Errorf("expected behavior serve-loop, got %q", co.Behavior)
	}
	if co.Reason != "tested via e2e binary spawn in tests/e2e/worker_test.rs" {
		t.Errorf("unexpected reason: %q", co.Reason)
	}
}

func TestLoadProjectCoveredMissingBehavior(t *testing.T) {
	path := writeSpec(t, `project: MyApp
sections:
  - name: Worker
    path: src/worker
    description: Worker process
    covered:
      - reason: tested via e2e
    behaviors:
      - name: serve-loop
        description: Main event loop
`)

	_, err := LoadProject(path)
	if err == nil {
		t.Error("expected error for covered override missing behavior")
	}
}

func TestLoadProjectCoveredMissingReason(t *testing.T) {
	path := writeSpec(t, `project: MyApp
sections:
  - name: Worker
    path: src/worker
    description: Worker process
    covered:
      - behavior: serve-loop
    behaviors:
      - name: serve-loop
        description: Main event loop
`)

	_, err := LoadProject(path)
	if err == nil {
		t.Error("expected error for covered override missing reason")
	}
}

func TestSectionChecksumIncludesCovered(t *testing.T) {
	sec := &Section{
		Name:        "Worker",
		Description: "worker",
		Behaviors:   []Behavior{{Name: "serve-loop", Description: "loop"}},
	}
	sum1 := SectionChecksum(sec, nil)

	sec.Covered = []CoveredOverride{{Behavior: "serve-loop", Reason: "e2e"}}
	sum2 := SectionChecksum(sec, nil)
	if sum1 == sum2 {
		t.Error("checksum should change when covered overrides are added")
	}
}

func TestLoadProjectWithDismissedOverrides(t *testing.T) {
	yml := `project: Test
sections:
  - name: Auth
    description: Auth module
    dismissed:
      - suggestion: logout
        reason: handled by session manager
    behaviors:
      - name: login
        description: POST /login
`
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.yaml")
	os.WriteFile(path, []byte(yml), 0644)

	ps, err := LoadProject(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(ps.Sections[0].Dismissed) != 1 {
		t.Fatalf("expected 1 dismissed override, got %d", len(ps.Sections[0].Dismissed))
	}
	d := ps.Sections[0].Dismissed[0]
	if d.Suggestion != "logout" || d.Reason != "handled by session manager" {
		t.Errorf("unexpected dismissed override: %+v", d)
	}
}

func TestLoadProjectDismissedMissingSuggestion(t *testing.T) {
	yml := `project: Test
sections:
  - name: Auth
    description: Auth module
    dismissed:
      - reason: no suggestion name
    behaviors:
      - name: login
        description: POST /login
`
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.yaml")
	os.WriteFile(path, []byte(yml), 0644)

	_, err := LoadProject(path)
	if err == nil {
		t.Error("expected error for dismissed override missing suggestion")
	}
}

func TestLoadProjectDismissedMissingReason(t *testing.T) {
	yml := `project: Test
sections:
  - name: Auth
    description: Auth module
    dismissed:
      - suggestion: logout
    behaviors:
      - name: login
        description: POST /login
`
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.yaml")
	os.WriteFile(path, []byte(yml), 0644)

	_, err := LoadProject(path)
	if err == nil {
		t.Error("expected error for dismissed override missing reason")
	}
}

func TestSectionChecksumIncludesDismissed(t *testing.T) {
	sec := &Section{
		Name:        "Auth",
		Description: "auth",
		Behaviors:   []Behavior{{Name: "login", Description: "login"}},
	}
	sum1 := SectionChecksum(sec, nil)

	sec.Dismissed = []DismissedOverride{{Suggestion: "logout", Reason: "skip"}}
	sum2 := SectionChecksum(sec, nil)
	if sum1 == sum2 {
		t.Error("checksum should change when dismissed overrides are added")
	}
}

func TestOversizedSections(t *testing.T) {
	// Build a section with 11 behaviors (over the limit)
	var behaviors []Behavior
	for i := 0; i < 11; i++ {
		behaviors = append(behaviors, Behavior{
			Name:        fmt.Sprintf("b%d", i),
			Description: fmt.Sprintf("behavior %d", i),
		})
	}

	ps := &ProjectSpec{
		Project: "Test",
		Sections: []Section{
			{Name: "Small", Description: "ok", Behaviors: []Behavior{{Name: "b", Description: "d"}}},
			{Name: "Big", Description: "too many", Behaviors: behaviors},
		},
	}

	oversized := ps.OversizedSections()
	if len(oversized) != 1 {
		t.Fatalf("expected 1 oversized section, got %d", len(oversized))
	}
	if oversized[0] != "Big" {
		t.Errorf("expected Big, got %s", oversized[0])
	}
}

func TestOversizedSectionsCountsSubsections(t *testing.T) {
	ps := &ProjectSpec{
		Project: "Test",
		Sections: []Section{
			{
				Name:        "Split",
				Description: "spread across subsections",
				Behaviors:   []Behavior{{Name: "b1", Description: "d1"}, {Name: "b2", Description: "d2"}},
				Subsections: []Subsection{
					{Name: "Sub", Behaviors: make([]Behavior, 9)},
				},
			},
		},
	}
	// 2 + 9 = 11 > 10
	for i := range ps.Sections[0].Subsections[0].Behaviors {
		ps.Sections[0].Subsections[0].Behaviors[i] = Behavior{
			Name:        fmt.Sprintf("s%d", i),
			Description: fmt.Sprintf("sub behavior %d", i),
		}
	}

	oversized := ps.OversizedSections()
	if len(oversized) != 1 {
		t.Fatalf("expected 1 oversized section, got %d", len(oversized))
	}
}

func TestOversizedSectionsNone(t *testing.T) {
	ps := &ProjectSpec{
		Project: "Test",
		Sections: []Section{
			{Name: "A", Description: "ok", Behaviors: []Behavior{{Name: "b", Description: "d"}}},
		},
	}

	oversized := ps.OversizedSections()
	if len(oversized) != 0 {
		t.Errorf("expected no oversized sections, got %v", oversized)
	}
}

func TestSectionPathsDeduplication(t *testing.T) {
	sec := &Section{
		Path: PathList{"internal/auth", "internal/core"},
		Subsections: []Subsection{
			{
				Name: "Sub1",
				Path: PathList{"internal/auth"},
			},
			{
				Name: "Sub2",
				Path: PathList{"internal/core", "internal/api"},
			},
		},
	}

	paths := SectionPaths(sec)
	// internal/auth, internal/core, internal/api — no duplicates
	if len(paths) != 3 {
		t.Errorf("expected 3 deduplicated paths, got %d: %v", len(paths), paths)
	}
	expected := []string{"internal/auth", "internal/core", "internal/api"}
	for i, want := range expected {
		if i >= len(paths) || paths[i] != want {
			t.Errorf("expected paths[%d]=%q, got %v", i, want, paths)
		}
	}
}
