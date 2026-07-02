// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

// Package main implements the onecli generator binary.
package main

import (
	"fmt"
	"os"

	"github.com/9Ashwin/one-cli/internal/build"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "onecli",
		Short: "onecli — Agent-Native CLI generation framework",
		Long:  "onecli generates AI-agent-friendly CLIs from OpenAPI 3.x specifications.",
	}
	root.AddCommand(newInitCmd())
	root.AddCommand(newVersionCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print onecli version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("onecli", build.Version)
		},
	}
}
