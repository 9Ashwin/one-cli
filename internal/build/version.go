// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

// Package build holds version metadata injected at build time via -ldflags.
package build

// Version is the semantic version (or git describe) of the binary.
// Set via: -ldflags "-X github.com/9Ashwin/one-cli/internal/build.Version=v0.1.0"
var Version = "dev"

// Date is the build date in YYYY-MM-DD form.
var Date = "unknown"
