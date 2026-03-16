package diff

import (
	"testing"

	"github.com/nwiley/vex/internal/lang"
)

func TestFilterByLanguageGo(t *testing.T) {
	l := &lang.Language{
		Name:           "go",
		TestPatterns:   []string{"*_test.go"},
		SourcePatterns: []string{"*.go"},
	}

	files := []string{
		"/repo/main.go",
		"/repo/main_test.go",
		"/repo/handler.go",
		"/repo/README.md",
	}

	src, tests := FilterByLanguage(files, l)

	if len(src) != 2 {
		t.Errorf("expected 2 source files, got %d: %v", len(src), src)
	}
	if len(tests) != 1 {
		t.Errorf("expected 1 test file, got %d: %v", len(tests), tests)
	}
}

func TestFilterByLanguageTS(t *testing.T) {
	l := &lang.Language{
		Name:           "typescript",
		TestPatterns:   []string{"*.test.ts", "*.spec.ts"},
		SourcePatterns: []string{"*.ts"},
	}

	files := []string{
		"/repo/app.ts",
		"/repo/app.test.ts",
		"/repo/utils.spec.ts",
		"/repo/style.css",
	}

	src, tests := FilterByLanguage(files, l)

	if len(src) != 1 {
		t.Errorf("expected 1 source file, got %d: %v", len(src), src)
	}
	if len(tests) != 2 {
		t.Errorf("expected 2 test files, got %d: %v", len(tests), tests)
	}
}

func TestFilterByLanguageNoMatches(t *testing.T) {
	l := &lang.Language{
		Name:           "go",
		TestPatterns:   []string{"*_test.go"},
		SourcePatterns: []string{"*.go"},
	}

	files := []string{"/repo/README.md", "/repo/Makefile"}

	src, tests := FilterByLanguage(files, l)

	if len(src) != 0 {
		t.Errorf("expected 0 source files, got %d", len(src))
	}
	if len(tests) != 0 {
		t.Errorf("expected 0 test files, got %d", len(tests))
	}
}

func TestChangedFilesNonGitDir(t *testing.T) {
	dir := t.TempDir()
	_, err := ChangedFiles(dir)
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}
