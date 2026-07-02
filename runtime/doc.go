// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

// Package runtime assembles the three-tier command architecture of a one-cli
// generated CLI.
//
// Tiers:
//
//  1. Shortcuts:        <cli> <service> +<shortcut>
//  2. Metadata-driven:  <cli> <service> <resource> <method>
//  3. Generic call:     <cli> api <METHOD> <path>
//
// The runtime wires metadata, auth, HTTP client, output formatting, and the
// typed error envelope together.
package runtime
