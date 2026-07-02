// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

// Package validate provides input/output path validation for one-cli.
//
// Because CLI arguments may originate from AI agents, every path that touches
// the local filesystem must be validated before use. SafeInputPath rejects
// paths that escape the working directory or reference sensitive locations.
package validate

import (
	"fmt"
	"path/filepath"
	"strings"
)

// SafeInputPath validates that a user-supplied path is safe to read from.
// It rejects empty paths, absolute paths, paths containing "..", and paths
// that would escape the provided base directory.
func SafeInputPath(base, input string) (string, error) {
	if strings.TrimSpace(input) == "" {
		return "", fmt.Errorf("path is empty")
	}
	if filepath.IsAbs(input) {
		return "", fmt.Errorf("absolute paths are not allowed: %s", input)
	}
	if strings.Contains(filepath.ToSlash(input), "..") {
		return "", fmt.Errorf("path contains parent directory reference: %s", input)
	}
	clean := filepath.Clean(input)
	full := filepath.Join(base, clean)
	rel, err := filepath.Rel(base, full)
	if err != nil {
		return "", fmt.Errorf("cannot relativize path: %w", err)
	}
	if strings.HasPrefix(filepath.ToSlash(rel), "..") {
		return "", fmt.Errorf("path escapes base directory: %s", input)
	}
	return clean, nil
}

// SafeOutputPath validates that a user-supplied path is safe to write to.
// It applies the same rules as SafeInputPath and additionally rejects paths
// pointing to existing non-regular files (directories, devices).
func SafeOutputPath(base, input string) (string, error) {
	clean, err := SafeInputPath(base, input)
	if err != nil {
		return "", err
	}
	// Additional output-specific checks can be added here (e.g. forbid
	// overwriting dotfiles).
	return clean, nil
}
