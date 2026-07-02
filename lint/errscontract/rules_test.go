// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package errscontract

import (
	"strings"
	"testing"
)

func TestCheckProblemEmbed_Missing(t *testing.T) {
	src := `package errs

type BadError struct {
	Message string
}
`
	vs := CheckProblemEmbed("errs/types.go", src)
	if len(vs) != 1 {
		t.Fatalf("want 1 violation, got %d", len(vs))
	}
	if vs[0].Rule != "problem_embed" {
		t.Errorf("rule = %q, want problem_embed", vs[0].Rule)
	}
}

func TestCheckProblemEmbed_Present(t *testing.T) {
	src := `package errs

type GoodError struct {
	Problem
	Cause error
}
`
	vs := CheckProblemEmbed("errs/types.go", src)
	if len(vs) != 0 {
		t.Fatalf("want 0 violations, got %d", len(vs))
	}
}

func TestCheckTypedErrorCompleteness_MissingCategory(t *testing.T) {
	src := `package main

import "github.com/9Ashwin/one-cli/errs"

func f() *errs.ValidationError {
	return &errs.ValidationError{
		Problem: errs.Problem{
			Subtype: errs.SubtypeInvalidArgument,
			Message: "x",
		},
	}
}
`
	vs := CheckTypedErrorCompleteness("cmd/foo.go", src)
	if len(vs) != 1 || vs[0].Rule != "typed_error_completeness" {
		t.Fatalf("want 1 typed_error_completeness violation, got %+v", vs)
	}
	if !strings.Contains(vs[0].Message, "Category") {
		t.Errorf("missing Category: %s", vs[0].Message)
	}
}

func TestCheckDeclaredSubtype_UndeclaredLiteral(t *testing.T) {
	src := `package main

import "github.com/9Ashwin/one-cli/errs"

func f() *errs.ValidationError {
	return &errs.ValidationError{
		Problem: errs.Problem{
			Category: errs.CategoryValidation,
			Subtype:  "totally_bogus",
			Message:  "x",
		},
	}
}
`
	allowlist := map[string]struct{}{"invalid_argument": {}}
	vs := CheckDeclaredSubtype("cmd/foo.go", src, allowlist)
	if len(vs) != 1 {
		t.Fatalf("want 1 violation, got %+v", vs)
	}
	if vs[0].Rule != "declared_subtype" || vs[0].Action != ActionReject {
		t.Errorf("want declared_subtype REJECT, got %+v", vs[0])
	}
}

func TestCheckDeclaredSubtype_DeclaredSelector(t *testing.T) {
	src := `package main

import "github.com/9Ashwin/one-cli/errs"

func f() *errs.ValidationError {
	return &errs.ValidationError{
		Problem: errs.Problem{
			Category: errs.CategoryValidation,
			Subtype:  errs.SubtypeInvalidArgument,
			Message:  "x",
		},
	}
}
`
	allowlist := map[string]struct{}{"invalid_argument": {}}
	vs := CheckDeclaredSubtype("cmd/foo.go", src, allowlist)
	if len(vs) != 0 {
		t.Fatalf("want 0 violations, got %+v", vs)
	}
}

func TestCheckNoBareCommandError_Reject(t *testing.T) {
	src := `package cmd

import "fmt"

var rootCmd = &cobra.Command{
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("bare error")
	},
}
`
	vs := CheckNoBareCommandError("cmd/root.go", src)
	if len(vs) != 1 {
		t.Fatalf("want 1 violation, got %+v", vs)
	}
	if vs[0].Rule != "no_bare_command_error" {
		t.Errorf("rule = %q, want no_bare_command_error", vs[0].Rule)
	}
}

func TestCheckNoBareCommandError_OutsideScope(t *testing.T) {
	src := `package internal

import "fmt"

func helper() error {
	return fmt.Errorf("intermediate wrap")
}
`
	vs := CheckNoBareCommandError("internal/helper.go", src)
	if len(vs) != 0 {
		t.Fatalf("want 0 violations, got %+v", vs)
	}
}

func TestScanRepo_NoViolationsOnSelf(t *testing.T) {
	vs, err := ScanRepo("../../")
	if err != nil {
		t.Fatalf("ScanRepo: %v", err)
	}
	var rejects []Violation
	for _, v := range vs {
		if v.Action == ActionReject {
			rejects = append(rejects, v)
		}
	}
	if len(rejects) != 0 {
		for _, v := range rejects {
			t.Errorf("unexpected REJECT: %s:%d %s/%s %s", v.File, v.Line, v.Action, v.Rule, v.Message)
		}
	}
}
