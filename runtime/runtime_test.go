// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package runtime

import (
	"bytes"
	"context"
	"testing"

	"github.com/9Ashwin/one-cli/auth"
	"github.com/9Ashwin/one-cli/client"
	"github.com/9Ashwin/one-cli/metadata"
	"github.com/9Ashwin/one-cli/output"
	"github.com/9Ashwin/one-cli/vfs"
)

func TestNew_RequiresMeta(t *testing.T) {
	_, err := New(Deps{})
	if err == nil {
		t.Fatal("expected error for nil metadata")
	}
}

func TestNew_RequiresServices(t *testing.T) {
	_, err := New(Deps{Meta: &metadata.Spec{CLIName: "x"}})
	if err == nil {
		t.Fatal("expected error for empty services")
	}
}

func TestRuntime_APICmd_DryRun(t *testing.T) {
	fake := &client.Fake{}
	r, err := New(Deps{
		Meta: &metadata.Spec{
			CLIName:         "petstore-cli",
			DefaultIdentity: auth.IdentityUser,
			Services: map[string]metadata.Service{
				"pets": {Name: "pets"},
			},
		},
		Auth:   auth.Map{auth.IdentityUser: {AccessToken: "tok", Identity: auth.IdentityUser}},
		Client: fake,
		Output: output.New(output.FormatJSON),
		VFS:    vfs.NewOS(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	var buf bytes.Buffer
	root := r.RootCmd()
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"api", "GET", "/pets", "--dry-run", "--format=json"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(fake.Requests) != 0 {
		t.Errorf("expected no HTTP request in dry-run, got %d", len(fake.Requests))
	}
	if !bytes.Contains(buf.Bytes(), []byte(`"dry_run":true`)) {
		t.Errorf("output missing dry_run marker: %s", buf.String())
	}
}

func TestRuntime_APICmd_RequiresLeadingSlash(t *testing.T) {
	r, err := New(Deps{
		Meta: &metadata.Spec{
			CLIName:         "petstore-cli",
			DefaultIdentity: auth.IdentityUser,
			Services: map[string]metadata.Service{
				"pets": {Name: "pets"},
			},
		},
		Client: &client.Fake{},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	root := r.RootCmd()
	root.SetArgs([]string{"api", "GET", "pets"})
	if err := root.ExecuteContext(context.Background()); err == nil {
		t.Fatal("expected validation error for path without leading slash")
	}
}

func TestRuntime_MethodCmd(t *testing.T) {
	fake := &client.Fake{Response: &client.Response{StatusCode: 200, Body: []byte(`{"items":[]}`)}}
	r, err := New(Deps{
		Meta: &metadata.Spec{
			CLIName:         "petstore-cli",
			DefaultIdentity: auth.IdentityUser,
			Services: map[string]metadata.Service{
				"pets": {
					Name: "pets",
					Resources: map[string]metadata.Resource{
						"pet": {
							Name: "pet",
							Methods: map[string]metadata.Method{
								"listPets": {
									OperationID: "listPets",
									HTTPMethod:  "GET",
									Path:        "/pets",
									Summary:     "List pets",
								},
							},
						},
					},
				},
			},
		},
		Auth:   auth.Map{auth.IdentityUser: {AccessToken: "tok", Identity: auth.IdentityUser}},
		Client: fake,
		Output: output.New(output.FormatJSON),
		VFS:    vfs.NewOS(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	var buf bytes.Buffer
	root := r.RootCmd()
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"pets", "pet", "listPets"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(fake.Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(fake.Requests))
	}
	if fake.Requests[0].Method != "GET" || fake.Requests[0].Path != "/pets" {
		t.Errorf("request = %s %s, want GET /pets", fake.Requests[0].Method, fake.Requests[0].Path)
	}
}

func TestRuntime_ShortcutCmd(t *testing.T) {
	fake := &client.Fake{Response: &client.Response{StatusCode: 200, Body: []byte(`{"items":[]}`)}}
	r, err := New(Deps{
		Meta: &metadata.Spec{
			CLIName:         "petstore-cli",
			DefaultIdentity: auth.IdentityUser,
			Services: map[string]metadata.Service{
				"pets": {
					Name: "pets",
					Resources: map[string]metadata.Resource{
						"pet": {
							Name: "pet",
							Methods: map[string]metadata.Method{
								"listPets": {
									OperationID: "listPets",
									HTTPMethod:  "GET",
									Path:        "/pets",
									Shortcut:    "+list",
									Summary:     "List pets",
								},
							},
						},
					},
				},
			},
		},
		Auth:   auth.Map{auth.IdentityUser: {AccessToken: "tok", Identity: auth.IdentityUser}},
		Client: fake,
		Output: output.New(output.FormatJSON),
		VFS:    vfs.NewOS(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	var buf bytes.Buffer
	root := r.RootCmd()
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"pets", "+list"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(fake.Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(fake.Requests))
	}
	if fake.Requests[0].Method != "GET" || fake.Requests[0].Path != "/pets" {
		t.Errorf("request = %s %s, want GET /pets", fake.Requests[0].Method, fake.Requests[0].Path)
	}
}
