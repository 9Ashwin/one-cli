// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package security

import "testing"

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"hello", "hello"},
		{"  hello  ", "hello"},
		{"he\x00llo", "hello"},
		{"he\x1fllo", "hello"},
	}
	for _, tt := range tests {
		if got := SanitizeString(tt.in); got != tt.want {
			t.Errorf("SanitizeString(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
