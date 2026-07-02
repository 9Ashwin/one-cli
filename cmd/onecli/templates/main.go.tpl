// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"

	"github.com/9Ashwin/one-cli/auth"
	"github.com/9Ashwin/one-cli/client"
	"github.com/9Ashwin/one-cli/metadata"
	"github.com/9Ashwin/one-cli/output"
	"github.com/9Ashwin/one-cli/runtime"
	"github.com/9Ashwin/one-cli/vfs"
)

//go:generate go run github.com/9Ashwin/one-cli/cmd/onecli init --spec ../../openapi.json --name {{.Name}}

func main() {
	deps := runtime.Deps{
		Meta: loadMeta(),
		Auth: auth.Map{
			auth.IdentityUser: {AccessToken: os.Getenv("{{.Name}}_TOKEN"), Identity: auth.IdentityUser},
		},
		Client: &client.HTTP{BaseURL: os.Getenv("{{.Name}}_BASE_URL")},
		Output: output.New(output.FormatPretty),
		VFS:    vfs.NewOS(),
	}
	r, err := runtime.New(deps)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := r.RootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadMeta() *metadata.Spec {
	return &metadata.Spec{
		CLIName:         "{{.Name}}",
		DefaultIdentity: auth.{{.DefaultIdentity}},
		Services:        map[string]metadata.Service{},
	}
}
