// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package build

import "testing"

func TestVersionDefaults(t *testing.T) {
	// Without -ldflags injection, Version defaults to "dev" and Date to "unknown".
	// This guards the contract that the binary always has a printable version.
	if Version == "" {
		t.Fatal("Version must never be empty; default is 'dev'")
	}
	if Date == "" {
		t.Fatal("Date must never be empty; default is 'unknown'")
	}
}
