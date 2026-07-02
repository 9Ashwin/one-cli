// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package metadata

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestFromOpenAPI_Basic(t *testing.T) {
	doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info:    &openapi3.Info{Title: "Petstore", Version: "1.0.0"},
		Paths:   &openapi3.Paths{},
	}
	doc.Paths.Set("/pets", &openapi3.PathItem{
		Get: &openapi3.Operation{
			OperationID: "listPets",
			Summary:     "List pets",
			Extensions: map[string]any{
				extShortcut: "+list",
			},
		},
		Post: &openapi3.Operation{
			OperationID: "createPet",
			Summary:     "Create a pet",
			Extensions: map[string]any{
				extQuality:    string(QualityStable),
				extDryRunSafe: true,
			},
		},
	})
	doc.Paths.Set("/pets/{id}", &openapi3.PathItem{
		Get: &openapi3.Operation{
			OperationID: "getPet",
			Summary:     "Get a pet",
			Parameters: openapi3.Parameters{
				{
					Value: &openapi3.Parameter{
						Name:     "id",
						In:       "path",
						Required: true,
						Schema:   &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
					},
				},
			},
		},
	})

	spec, err := FromOpenAPI("petstore-cli", doc)
	if err != nil {
		t.Fatalf("FromOpenAPI: %v", err)
	}
	if spec.CLIName != "petstore-cli" {
		t.Errorf("CLIName = %q, want petstore-cli", spec.CLIName)
	}
	svc, ok := spec.Services["pets"]
	if !ok {
		t.Fatalf("missing pets service, got %v", spec.Services)
	}
	res, ok := svc.Resources["pet"]
	if !ok {
		t.Fatalf("missing pet resource")
	}
	if len(res.Methods) != 3 {
		t.Errorf("methods = %d, want 3", len(res.Methods))
	}
	m, ok := res.Methods["listPets"]
	if !ok {
		t.Fatalf("missing listPets method")
	}
	if m.Shortcut != "+list" {
		t.Errorf("Shortcut = %q, want +list", m.Shortcut)
	}
	if m.Quality != QualityExperimental {
		t.Errorf("Quality = %q, want experimental", m.Quality)
	}

	post, ok := res.Methods["createPet"]
	if !ok {
		t.Fatalf("missing createPet method")
	}
	if post.Quality != QualityStable {
		t.Errorf("createPet Quality = %q, want stable", post.Quality)
	}
	if !post.DryRunSafe {
		t.Errorf("createPet DryRunSafe = false, want true")
	}
}

func TestFromOpenAPI_Extensions(t *testing.T) {
	doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info:    &openapi3.Info{Title: "X", Version: "1"},
		Paths:   &openapi3.Paths{},
	}
	doc.Paths.Set("/things", &openapi3.PathItem{
		Get: &openapi3.Operation{
			OperationID: "doThing",
			Extensions: map[string]any{
				extIdentities: []any{"user", "bot"},
				extScopes:     []any{"thing:read"},
			},
		},
	})
	spec, err := FromOpenAPI("x-cli", doc)
	if err != nil {
		t.Fatalf("FromOpenAPI: %v", err)
	}
	svc := spec.Services["things"]
	res := svc.Resources["thing"]
	m := res.Methods["doThing"]
	if len(m.Identities) != 2 {
		t.Errorf("Identities = %v, want 2", m.Identities)
	}
	if len(m.Scopes) != 1 || m.Scopes[0] != "thing:read" {
		t.Errorf("Scopes = %v, want [thing:read]", m.Scopes)
	}
}

func TestFromOpenAPI_EmptyPaths(t *testing.T) {
	doc := &openapi3.T{OpenAPI: "3.0.0", Info: &openapi3.Info{Title: "Empty"}}
	_, err := FromOpenAPI("empty-cli", doc)
	if err == nil {
		t.Fatal("expected error for empty paths")
	}
}
