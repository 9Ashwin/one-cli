// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package errscontract

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"
)

// CheckDeclaredSubtype enforces that `Subtype:` literals resolve to a
// declared constant value (allowlist), match the ad_hoc_* namespace
// (deferred to CheckAdHocSubtype), or are dynamic (WARNING). Undeclared
// static literals are rejected.
//
// allowlist holds declared Subtype const values (e.g. "missing_scope"). The
// production CLI derives this from errs/subtypes*.go via the AST; unit tests
// pass in a fixture map. Passing nil disables CheckDeclaredSubtype entirely.
func CheckDeclaredSubtype(path, src string, allowlist map[string]struct{}) []Violation {
	if allowlist == nil {
		return nil
	}
	v := scanSubtype(path, src, allowlist)
	out := v[:0]
	for _, vv := range v {
		if vv.Rule == "declared_subtype" {
			out = append(out, vv)
		}
	}
	return out
}

// CheckAdHocSubtype flags ad_hoc_* Subtype literals with a LABEL action so
// CI can mark them for follow-up taxonomy decisions.
func CheckAdHocSubtype(path, src string) []Violation {
	v := scanSubtype(path, src, nil)
	out := v[:0]
	for _, vv := range v {
		if vv.Rule == "adhoc_subtype" {
			out = append(out, vv)
		}
	}
	return out
}

// scanSubtype walks the file AST and classifies every `Subtype:` key-value
// assignment in a composite literal.
func scanSubtype(path, src string, allowlist map[string]struct{}) []Violation {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return nil
	}
	adHoc := regexp.MustCompile(`^ad_hoc_[a-z0-9_]+$`)
	var out []Violation
	ast.Inspect(file, func(n ast.Node) bool {
		cl, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		for _, el := range cl.Elts {
			kv, ok := el.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			keyIdent, ok := kv.Key.(*ast.Ident)
			if !ok || keyIdent.Name != "Subtype" {
				continue
			}
			if v := classifySubtypeExpr(kv.Value, allowlist, adHoc); v.Rule != "" {
				v.File = path
				v.Line = fset.Position(kv.Pos()).Line
				out = append(out, v)
			}
		}
		return true
	})
	return out
}

// classifySubtypeExpr inspects a single expression sitting in a `Subtype:`
// slot and returns the lint verdict.
func classifySubtypeExpr(expr ast.Expr, allowlist map[string]struct{}, adHoc *regexp.Regexp) Violation {
	switch v := expr.(type) {
	case *ast.SelectorExpr:
		if v.Sel == nil {
			return Violation{}
		}
		// Only Subtype-prefixed selector names are treated as constant refs.
		if !isSubtypeConstName(v.Sel.Name) {
			return Violation{}
		}
		return Violation{}
	case *ast.Ident:
		if isSubtypeConstName(v.Name) {
			return Violation{}
		}
		return Violation{
			Rule:       "declared_subtype",
			Action:     ActionWarning,
			Message:    "Subtype assigned from identifier " + v.Name + " — value resolution requires manual review",
			Suggestion: "prefer named constants from errs/subtypes.go (e.g. errs.SubtypeMissingScope); if dynamic, justify in PR description",
		}
	case *ast.BasicLit:
		if v.Kind != token.STRING {
			return Violation{}
		}
		return classifyStringValue(unquoteSimple(v.Value), allowlist, adHoc)
	case *ast.CallExpr:
		if !isSubtypeCast(v.Fun) || len(v.Args) != 1 {
			return Violation{}
		}
		lit, ok := v.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return Violation{
				Rule:       "declared_subtype",
				Action:     ActionWarning,
				Message:    "errs.Subtype(...) cast from non-literal expression — value resolution requires manual review",
				Suggestion: "prefer named constants from errs/subtypes.go",
			}
		}
		return classifySubtypeExpr(lit, allowlist, adHoc)
	}
	return Violation{}
}

func isSubtypeConstName(name string) bool {
	return strings.HasPrefix(name, "Subtype") && name != "Subtype"
}

func isSubtypeCast(fun ast.Expr) bool {
	switch f := fun.(type) {
	case *ast.Ident:
		return f.Name == "Subtype"
	case *ast.SelectorExpr:
		return f.Sel != nil && f.Sel.Name == "Subtype"
	}
	return false
}

func classifyStringValue(value string, allowlist map[string]struct{}, adHoc *regexp.Regexp) Violation {
	if adHoc != nil && adHoc.MatchString(value) {
		return Violation{
			Rule:       "adhoc_subtype",
			Action:     ActionLabel,
			Message:    `Subtype "` + value + `" matches ad_hoc_* temporary namespace — add label "needs-taxonomy-decision" [needs-taxonomy-decision]`,
			Suggestion: "promote ad_hoc_* to a declared Subtype constant within 1 week",
		}
	}
	if allowlist == nil {
		return Violation{}
	}
	if _, ok := allowlist[value]; ok {
		return Violation{}
	}
	return Violation{
		Rule:    "declared_subtype",
		Action:  ActionReject,
		Message: `Subtype "` + value + `" is not declared in errs/subtypes.go and does not match ad_hoc_* namespace`,
		Suggestion: "use a declared const from errs/subtypes.go (e.g. errs.SubtypeMissingScope), " +
			"or use ad_hoc_<name> temporarily and file a taxonomy issue",
	}
}

// unquoteSimple strips one layer of surrounding double or back quotes.
func unquoteSimple(quoted string) string {
	if len(quoted) >= 2 && (quoted[0] == '"' || quoted[0] == '`') {
		return quoted[1 : len(quoted)-1]
	}
	return quoted
}
