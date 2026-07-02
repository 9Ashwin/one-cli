// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunInit_GeneratesFiles(t *testing.T) {
	out := t.TempDir()
	spec := []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "Petstore", "version": "1.0.0"},
		"paths": {
			"/pets": {
				"get": {
					"operationId": "listPets",
					"summary": "List pets"
				}
			}
		}
	}`)
	specPath := filepath.Join(t.TempDir(), "openapi.json")
	if err := os.WriteFile(specPath, spec, 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	if err := runInit(specPath, "petstore-cli", out, true); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	for _, file := range []string{"main.go", "go.mod", "cmd/root.go", "README.md"} {
		full := filepath.Join(out, file)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("missing generated file %s: %v", file, err)
		}
	}
}

func TestRunInit_RequiresSpec(t *testing.T) {
	out := t.TempDir()
	if err := runInit("/nonexistent/openapi.json", "x-cli", out, true); err == nil {
		t.Fatal("expected error for missing spec")
	}
}

func TestRender(t *testing.T) {
	got := render("hello {{.Name}}", genContext{Name: "world"})
	if got != "hello world" {
		t.Errorf("render = %q, want hello world", got)
	}
}
