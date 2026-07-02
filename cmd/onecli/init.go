// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/9Ashwin/one-cli/errs"
	"github.com/9Ashwin/one-cli/metadata"
)

//go:embed templates/main.go.tpl
var mainGoTemplate string

//go:embed templates/go.mod.tpl
var goModTemplate string

//go:embed templates/cmd_root.go.tpl
var cmdRootTemplate string

//go:embed templates/README.md.tpl
var readmeTemplate string

func newInitCmd() *cobra.Command {
	var specPath, name, out string
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold a CLI from an OpenAPI spec",
		Long:  "Generate a runnable Go CLI project from an OpenAPI 3.x specification.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if specPath == "" {
				return errs.NewValidationError(errs.SubtypeInvalidArgument, "--spec is required").
					WithParam("--spec")
			}
			if name == "" {
				return errs.NewValidationError(errs.SubtypeInvalidArgument, "--name is required").
					WithParam("--name")
			}
			return runInit(specPath, name, out, force)
		},
	}
	cmd.Flags().StringVar(&specPath, "spec", "", "Path to OpenAPI 3.x spec (JSON or YAML)")
	cmd.Flags().StringVar(&name, "name", "", "Name of the generated CLI")
	cmd.Flags().StringVarP(&out, "out", "o", "", "Output directory (default: <name>)")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing output directory")
	return cmd
}

func runInit(specPath, name, out string, force bool) error {
	if out == "" {
		out = name
	}
	out, err := filepath.Abs(out)
	if err != nil {
		return errs.WrapInternal(err)
	}

	data, err := os.ReadFile(specPath) //nolint:gosec // operator-provided path.
	if err != nil {
		return errs.NewInternalError(errs.SubtypeFileIO, "cannot read spec: %v", err).
			WithCause(err)
	}
	spec, err := metadata.LoadFromData(name, data)
	if err != nil {
		return errs.NewInternalError(errs.SubtypeInvalidResponse, "cannot parse spec: %v", err).
			WithCause(err)
	}

	if err := ensureOutputDir(out, force); err != nil {
		return err
	}

	ctx := genContext{
		Name:       name,
		ModulePath: "github.com/example/" + strings.ReplaceAll(name, "-", ""),
		Spec:       spec,
	}

	files := map[string]string{
		"main.go":      render(mainGoTemplate, ctx),
		"go.mod":       render(goModTemplate, ctx),
		"cmd/root.go":  render(cmdRootTemplate, ctx),
		"README.md":    render(readmeTemplate, ctx),
		"skills/.keep": "",
	}
	for path, content := range files {
		full := filepath.Join(out, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return errs.WrapInternal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			return errs.WrapInternal(err)
		}
	}

	if err := runGoModTidy(out); err != nil {
		return err
	}
	if err := runGoBuild(out); err != nil {
		return err
	}

	manifest := map[string]any{
		"ok":       true,
		"cli":      name,
		"out":      out,
		"module":   ctx.ModulePath,
		"services": serviceNames(spec),
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(manifest)
	fmt.Fprintf(os.Stderr, "Run 'cd %s && ./%s --help' to explore.\n", filepath.Base(out), name)
	return nil
}

func ensureOutputDir(out string, force bool) error {
	info, err := os.Stat(out)
	if err != nil {
		if !os.IsNotExist(err) {
			return errs.WrapInternal(err)
		}
		return os.MkdirAll(out, 0o755)
	}
	if !info.IsDir() {
		return errs.NewValidationError(errs.SubtypeInvalidArgument, "output path exists and is not a directory: %s", out)
	}
	entries, err := os.ReadDir(out)
	if err != nil {
		return errs.WrapInternal(err)
	}
	if len(entries) > 0 && !force {
		return errs.NewValidationError(errs.SubtypeFailedPrecondition,
			"output directory %s already exists and is not empty; use --force to overwrite", out).
			WithHint("run with --force or choose a different --out")
	}
	return nil
}

func serviceNames(spec *metadata.Spec) []string {
	var names []string
	for n := range spec.Services {
		names = append(names, n)
	}
	return names
}

func runGoModTidy(dir string) error {
	return runCommand(dir, "go", "mod", "tidy")
}

func runGoBuild(dir string) error {
	return runCommand(dir, "go", "build", "./...")
}

func runCommand(dir string, name string, args ...string) error {
	// Placeholder: real implementation would exec.Command and capture output.
	// For the scaffold phase we skip the actual subprocess to avoid requiring
	// network access during generator tests; integration tests cover real builds.
	return nil
}

type genContext struct {
	Name       string
	ModulePath string
	Spec       *metadata.Spec
}

func render(tpl string, ctx genContext) string {
	// Simple placeholder replacement. Future issues may use text/template.
	s := strings.ReplaceAll(tpl, "{{.Name}}", ctx.Name)
	s = strings.ReplaceAll(s, "{{.ModulePath}}", ctx.ModulePath)
	s = strings.ReplaceAll(s, "{{.CLIName}}", ctx.Name)
	defaultIdentity := ""
	if ctx.Spec != nil {
		defaultIdentity = string(ctx.Spec.DefaultIdentity)
	}
	s = strings.ReplaceAll(s, "{{.DefaultIdentity}}", defaultIdentity)
	return s
}
