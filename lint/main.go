// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

// Command lintcheck runs the source-level errs/ contract guards.
//
// lintcheck lives in its own Go module under lint/ so its build-time
// dependency on go/ast does not leak into the shipped onecli binary's
// module graph.
//
// Usage (from repo root):
//
//	go run -C lint . .                # scan the one-cli repo
//	go run -C lint . /path/to/repo    # scan another path
//
// Exit codes:
//
//	0  no REJECT violations (LABEL and WARNING diagnostics are advisory)
//	1  one or more REJECT violations
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/9Ashwin/one-cli/lint/errscontract"
	"github.com/9Ashwin/one-cli/lint/lintapi"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"Usage: lintcheck [repo-root]\n"+
				"Runs errscontract checks against repo-root (default: current directory).\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	root := "."
	if flag.NArg() > 0 {
		root = flag.Arg(0)
		if root == "./..." {
			root = "."
		}
	}

	violations, err := errscontract.ScanRepo(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lintcheck errscontract: %v\n", err)
		os.Exit(2)
	}

	exitCode := 0
	for _, v := range violations {
		fmt.Fprintf(os.Stderr, "%s:%d: [%s/%s] %s\n", v.File, v.Line, v.Action, v.Rule, v.Message)
		if v.Suggestion != "" {
			fmt.Fprintf(os.Stderr, "    hint: %s\n", v.Suggestion)
		}
		if v.Action == lintapi.ActionReject {
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}
