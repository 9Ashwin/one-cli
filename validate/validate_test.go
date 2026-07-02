// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package validate

import (
	"path/filepath"
	"testing"
)

func TestSafeInputPath(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		input   string
		wantErr bool
	}{
		{"simple", "/tmp", "file.txt", false},
		{"nested", "/tmp", "dir/file.txt", false},
		{"empty", "/tmp", "", true},
		{"absolute", "/tmp", "/etc/passwd", true},
		{"parent escape", "/tmp", "../etc/passwd", true},
		{"dotdot inside", "/tmp", "foo/../bar", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeInputPath(tt.base, tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SafeInputPath(%q,%q) err=%v, wantErr=%v", tt.base, tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got == "" {
				t.Errorf("expected non-empty clean path")
			}
		})
	}
}

func TestSafeInputPath_Relativize(t *testing.T) {
	base := filepath.FromSlash("/tmp/wd")
	input := "data/file.txt"
	got, err := SafeInputPath(base, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "data/file.txt" {
		t.Errorf("got %q, want %q", got, "data/file.txt")
	}
}
