// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package errscontract

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ScanRepo walks the repo rooted at root and emits violations covering the
// errscontract checks. It derives the Subtype allowlist from errs/subtypes*.go.
func ScanRepo(root string) ([]Violation, error) {
	allowlist, err := loadSubtypeAllowlist(filepath.Join(root, "errs"))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("load subtype allowlist: %w", err)
		}
		allowlist = nil
	}

	var all []Violation
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if path != root && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			if d.Name() == "vendor" || d.Name() == "node_modules" || d.Name() == "docs" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		src, err := os.ReadFile(path) //nolint:gosec // CLI tool; root is operator-provided.
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		all = append(all, CheckNoBareCommandError(rel, string(src))...)
		all = append(all, CheckAdHocSubtype(rel, string(src))...)
		all = append(all, CheckTypedErrorCompleteness(rel, string(src))...)
		if allowlist != nil && !isErrsScope(rel) {
			all = append(all, CheckDeclaredSubtype(rel, string(src), allowlist)...)
		}
		if isErrsScope(rel) {
			all = append(all, CheckProblemEmbed(rel, string(src))...)
		}
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	sort.SliceStable(all, func(i, j int) bool {
		if all[i].File != all[j].File {
			return all[i].File < all[j].File
		}
		return all[i].Line < all[j].Line
	})
	return all, nil
}

// loadSubtypeAllowlist parses every errs/subtypes*.go file under dir and
// returns the set of declared Subtype constant VALUES.
func loadSubtypeAllowlist(errsDir string) (map[string]struct{}, error) {
	if _, statErr := os.Stat(errsDir); statErr != nil {
		return nil, statErr
	}
	entries, err := os.ReadDir(errsDir)
	if err != nil {
		return nil, err
	}
	allowlist := make(map[string]struct{})
	found := false
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "subtypes") || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		full := filepath.Join(errsDir, name)
		values, err := loadSubtypeAllowlistFile(full)
		if err != nil {
			return nil, err
		}
		for k := range values {
			allowlist[k] = struct{}{}
		}
		found = true
	}
	if !found {
		return nil, fmt.Errorf("%w: no subtypes*.go found under %s", os.ErrNotExist, errsDir)
	}
	return allowlist, nil
}

func loadSubtypeAllowlistFile(subtypesGo string) (map[string]struct{}, error) {
	src, err := os.ReadFile(subtypesGo) //nolint:gosec // operator-provided path.
	if err != nil {
		return nil, err
	}
	return parseSubtypeValues(string(src))
}

// isErrsScope reports whether a path is inside the errs/ package.
func isErrsScope(path string) bool {
	p := strings.ReplaceAll(path, "\\", "/")
	return strings.HasPrefix(p, "errs/") || strings.Contains(p, "/errs/")
}
