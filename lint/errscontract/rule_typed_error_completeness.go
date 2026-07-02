// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package errscontract

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// CheckTypedErrorCompleteness rejects typed `*errs.<X>Error` composite
// literals whose embedded Problem is missing any of the three required
// fields: Category, Subtype, Message. Without this check, new code can
// silently introduce typed errors that emit empty `type` / `subtype` on the
// wire and confuse downstream consumers.
//
// Fires only when:
//   - the type is a qualified `errs.<X>Error` selector, OR
//   - the file lives inside the canonical errs package and the type is an
//     unqualified `<X>Error` ident.
func CheckTypedErrorCompleteness(path, src string) []Violation {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return nil
	}
	inErrsPackage := isErrsPackagePath(path)
	var out []Violation
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		errorName, isErrsType := typedErrorTypeName(lit.Type, inErrsPackage)
		if !isErrsType {
			return true
		}
		problemLit, kind := findProblemLiteral(lit)
		switch kind {
		case problemMissing:
			out = append(out, completenessReject(path, fset.Position(lit.Pos()).Line, errorName, "Problem"))
		case problemLiteral:
			for _, required := range []string{"Category", "Subtype", "Message"} {
				if !hasKeyedEntry(problemLit, required) {
					out = append(out, completenessReject(path, fset.Position(problemLit.Pos()).Line, errorName, required))
				}
			}
		}
		return true
	})
	return out
}

// typedErrorTypeName reports whether a composite-literal Type names a typed
// *errs.XxxError struct, and returns the bare type name for the diagnostic.
func typedErrorTypeName(expr ast.Expr, inErrsPackage bool) (string, bool) {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		x, ok := t.X.(*ast.Ident)
		if !ok || x.Name != "errs" || t.Sel == nil {
			return "", false
		}
		return t.Sel.Name, strings.HasSuffix(t.Sel.Name, "Error") && t.Sel.Name != "Error"
	case *ast.Ident:
		if !inErrsPackage {
			return "", false
		}
		return t.Name, strings.HasSuffix(t.Name, "Error") && t.Name != "Error"
	}
	return "", false
}

// isErrsPackagePath reports whether the given file path is inside the
// canonical errs/ package (top-level errs/ files only).
func isErrsPackagePath(path string) bool {
	p := strings.ReplaceAll(path, "\\", "/")
	if !strings.HasPrefix(p, "errs/") && !strings.Contains(p, "/errs/") {
		return false
	}
	var rest string
	if i := strings.Index(p, "/errs/"); i >= 0 {
		rest = p[i+len("/errs/"):]
	} else {
		rest = p[len("errs/"):]
	}
	return !strings.Contains(rest, "/")
}

// problemKind is the verdict of findProblemLiteral.
type problemKind int

const (
	problemMissing  problemKind = iota // no Problem key in the outer literal
	problemVariable                    // Problem value is a variable / call expr
	problemLiteral                     // Problem value is itself a composite literal
)

// findProblemLiteral returns the inner Problem composite literal and a
// problemKind verdict.
func findProblemLiteral(outer *ast.CompositeLit) (*ast.CompositeLit, problemKind) {
	for _, el := range outer.Elts {
		kv, ok := el.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "Problem" {
			continue
		}
		inner, ok := kv.Value.(*ast.CompositeLit)
		if !ok {
			return nil, problemVariable
		}
		return inner, problemLiteral
	}
	return nil, problemMissing
}

// hasKeyedEntry reports whether a composite literal contains a `<key>:`
// keyed entry.
func hasKeyedEntry(lit *ast.CompositeLit, key string) bool {
	for _, el := range lit.Elts {
		kv, ok := el.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		ident, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		if ident.Name == key {
			return true
		}
	}
	return false
}

func completenessReject(path string, line int, errorName, missing string) Violation {
	return Violation{
		Rule:    "typed_error_completeness",
		Action:  ActionReject,
		File:    path,
		Line:    line,
		Message: "typed *" + errorName + " literal is missing required Problem." + missing + " field",
		Suggestion: "every typed *errs.XxxError must set Problem.Category, Problem.Subtype, and Problem.Message — " +
			"missing fields emit an empty `type` / `subtype` / `message` on the wire and confuse consumers",
	}
}
