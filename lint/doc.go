// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

// Command lintcheck runs one-cli's source-level errs/ contract guards.
//
// It lives in its own Go module (github.com/9Ashwin/one-cli/lint) so its
// build-time dependency on go/ast and related tooling does not leak into the
// shipped onecli binary's module graph.
//
// The errscontract analyzer enforces the typed error contract defined in the
// main module's errs/ package: every command boundary error must be a typed
// errs.* error, every Subtype must be declared, and every typed error struct
// must embed errs.Problem.
package main
