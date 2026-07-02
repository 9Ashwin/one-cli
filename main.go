// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

// Package main is the entry point for the onecli generator binary.
//
// one-cli is an Agent-Native CLI generation framework: it scaffolds an
// AI-agent-friendly CLI from an OpenAPI 3.x spec. This binary hosts the
// generator commands (init / skill gen / quality check). The framework SDK
// lives in the top-level exported packages (runtime, auth, errs, ...).
package main

import (
	"fmt"
	"os"

	"github.com/9Ashwin/one-cli/internal/build"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Println("onecli", build.Version)
		return
	}
	fmt.Fprintln(os.Stderr, "onecli — Agent-Native CLI generation framework")
	fmt.Fprintln(os.Stderr, "Run 'onecli --help' once subcommands are implemented.")
	fmt.Fprintln(os.Stderr, "See: https://github.com/9Ashwin/one-cli")
	os.Exit(0)
}
