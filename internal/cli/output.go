package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

const vexDir = ".vex"

func writeOutput(filename string, data []byte) error {
	if err := os.MkdirAll(vexDir, 0755); err != nil {
		return fmt.Errorf("creating %s directory: %w", vexDir, err)
	}

	path := filepath.Join(vexDir, filename)
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	fmt.Fprintln(os.Stderr, path)
	return nil
}
