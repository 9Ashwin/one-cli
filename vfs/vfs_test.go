// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package vfs

import (
	"path/filepath"
	"testing"
)

func TestMemVFS(t *testing.T) {
	base := t.TempDir()
	fsys := NewMem(base)
	if err := fsys.WriteFile("hello.txt", []byte("world"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	data, err := fsys.ReadFile("hello.txt")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "world" {
		t.Errorf("data = %q, want world", string(data))
	}
	if _, err := fsys.Stat(filepath.Join(base, "hello.txt")); err != nil {
		t.Errorf("Stat: %v", err)
	}
}

func TestOSVFS(t *testing.T) {
	fsys := NewOS()
	base := t.TempDir()
	path := filepath.Join(base, "x.txt")
	if err := fsys.WriteFile(path, []byte("os"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	data, err := fsys.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "os" {
		t.Errorf("data = %q, want os", string(data))
	}
}
