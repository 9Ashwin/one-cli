// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

// Package vfs provides a filesystem abstraction so generated CLIs and the
// one-cli generator can be tested without touching the real OS.
package vfs

import (
	"io/fs"
	"os"
	"path/filepath"
)

// VFS is the filesystem abstraction used by one-cli.
type VFS interface {
	// ReadFile returns the contents of a file.
	ReadFile(name string) ([]byte, error)
	// WriteFile writes data to a file, creating it if necessary.
	WriteFile(name string, data []byte, perm fs.FileMode) error
	// MkdirAll creates a directory and all necessary parents.
	MkdirAll(path string, perm fs.FileMode) error
	// Stat returns file info.
	Stat(name string) (fs.FileInfo, error)
	// Open opens a file for reading.
	Open(name string) (fs.File, error)
}

// OS is the real filesystem implementation.
type OS struct{}

// NewOS returns a VFS backed by the operating system.
func NewOS() VFS { return OS{} }

func (OS) ReadFile(name string) ([]byte, error)     { return os.ReadFile(name) }
func (OS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}
func (OS) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }
func (OS) Stat(name string) (fs.FileInfo, error)        { return os.Stat(name) }
func (OS) Open(name string) (fs.File, error)            { return os.Open(name) }

// Mem is an in-memory VFS for tests. It stores files relative to a base
// directory and delegates reads/writes to that directory on the real OS.
type Mem struct {
	Base string
}

// NewMem returns a VFS rooted at the given base directory.
func NewMem(base string) VFS { return Mem{Base: base} }

func (m Mem) resolve(name string) string {
	if filepath.IsAbs(name) {
		return name
	}
	return filepath.Join(m.Base, name)
}

func (m Mem) ReadFile(name string) ([]byte, error)     { return os.ReadFile(m.resolve(name)) }
func (m Mem) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(m.resolve(name), data, perm)
}
func (m Mem) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(m.resolve(path), perm) }
func (m Mem) Stat(name string) (fs.FileInfo, error)        { return os.Stat(m.resolve(name)) }
func (m Mem) Open(name string) (fs.File, error)            { return os.Open(m.resolve(name)) }
