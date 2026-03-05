package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

// resolveOutputPath resolves the output file path.
// If outPath is a directory, the defaultName is appended.
// The path is cleaned and converted to absolute form.
func resolveOutputPath(outPath, defaultName string) (string, error) {
	outPath = filepath.Clean(outPath)

	// Check if it's a directory.
	info, err := os.Stat(outPath)
	if err == nil && info.IsDir() {
		outPath = filepath.Join(outPath, defaultName)
	}

	abs, err := filepath.Abs(outPath)
	if err != nil {
		return "", fmt.Errorf("resolve path %q: %w", outPath, err)
	}
	return abs, nil
}

// resolveInputPath validates that the input file exists and returns its absolute path.
func resolveInputPath(inPath string) (string, error) {
	inPath = filepath.Clean(inPath)

	abs, err := filepath.Abs(inPath)
	if err != nil {
		return "", fmt.Errorf("resolve path %q: %w", inPath, err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("input file %q: %w", abs, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("input path %q is a directory, not a file", abs)
	}

	return abs, nil
}
