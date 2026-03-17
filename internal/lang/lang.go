package lang

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nickw409/vex/internal/config"
)

type Language struct {
	Name           string
	TestPatterns   []string
	SourcePatterns []string
}

var builtinLanguages = map[string]Language{
	"go": {
		Name:           "go",
		TestPatterns:   []string{"*_test.go"},
		SourcePatterns: []string{"*.go"},
	},
	"typescript": {
		Name:           "typescript",
		TestPatterns:   []string{"*.test.ts", "*.spec.ts"},
		SourcePatterns: []string{"*.ts"},
	},
	"javascript": {
		Name:           "javascript",
		TestPatterns:   []string{"*.test.js", "*.spec.js"},
		SourcePatterns: []string{"*.js"},
	},
	"python": {
		Name:           "python",
		TestPatterns:   []string{"test_*.py", "*_test.py"},
		SourcePatterns: []string{"*.py"},
	},
	"java": {
		Name:           "java",
		TestPatterns:   []string{"*Test.java"},
		SourcePatterns: []string{"*.java"},
	},
}

func Detect(dir string, overrides map[string]config.LanguageConfig) (*Language, error) {
	if len(overrides) > 0 {
		for name, lc := range overrides {
			return &Language{
				Name:           name,
				TestPatterns:   lc.TestPatterns,
				SourcePatterns: lc.SourcePatterns,
			}, nil
		}
	}

	counts := make(map[string]int)

	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := d.Name()
			if base == "node_modules" || base == "vendor" || base == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		name := d.Name()
		ext := strings.ToLower(filepath.Ext(name))

		switch ext {
		case ".go":
			counts["go"]++
		case ".ts":
			counts["typescript"]++
		case ".js":
			if _, err := os.Stat(filepath.Join(dir, "package.json")); err == nil {
				counts["javascript"]++
			}
		case ".py":
			counts["python"]++
		case ".java":
			counts["java"]++
		}
		return nil
	})

	if len(counts) == 0 {
		return nil, fmt.Errorf("no supported language detected in %s", dir)
	}

	best := ""
	bestCount := 0
	for lang, count := range counts {
		if count > bestCount {
			best = lang
			bestCount = count
		}
	}

	l := builtinLanguages[best]
	return &l, nil
}

func FindFiles(dir string, lang *Language) (sourceFiles []string, testFiles []string, err error) {
	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			base := d.Name()
			if base == "node_modules" || base == "vendor" || base == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		name := d.Name()

		if isTest(name, lang.TestPatterns) {
			testFiles = append(testFiles, path)
			return nil
		}

		if matchesAny(name, lang.SourcePatterns) {
			sourceFiles = append(sourceFiles, path)
		}

		return nil
	})

	return
}

// IsTestFile reports whether the given filename matches the language's test patterns.
func IsTestFile(filename string, lang *Language) bool {
	return matchesAny(filepath.Base(filename), lang.TestPatterns)
}

func isTest(name string, patterns []string) bool {
	return matchesAny(name, patterns)
}

func matchesAny(name string, patterns []string) bool {
	for _, p := range patterns {
		if matched, _ := filepath.Match(p, name); matched {
			return true
		}
	}
	return false
}
