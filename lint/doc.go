// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

// Package lint hosts one-cli's custom golangci-lint analyzers.
//
// It is an independent Go module (github.com/9Ashwin/one-cli/lint). The
// errscontract analyzer (banning bare fmt.Errorf final wraps, undeclared
// error subtypes, and legacy error helpers) is populated by Issue #11; this
// placeholder keeps the module buildable in the meantime.
package lint
