// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

// Package security provides output sanitization helpers. Because CLI arguments
// may come from AI agents, values written to logs, error messages, or shell
// output must be sanitized to avoid leaking secrets or injecting control
// characters.
package security

import (
	"regexp"
	"strings"
)

var controlChars = regexp.MustCompile(`[\x00-\x08\x0b-\x0c\x0e-\x1f\x7f]`)

// SanitizeString removes control characters and trims whitespace from a string.
// It is the default sanitizer for values that may appear in CLI output.
func SanitizeString(s string) string {
	s = strings.TrimSpace(s)
	s = controlChars.ReplaceAllString(s, "")
	return s
}

// SanitizePath cleans a path string for display in error messages. It does NOT
// validate the path; use validate.SafeInputPath for validation.
func SanitizePath(s string) string {
	return SanitizeString(s)
}
