// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package errscontract

import (
	"go/ast"
	"go/parser"
	"go/token"
)

// parseSubtypeValues extracts the string VALUES of all typed Subtype
// constants from a subtypes*.go source file.
func parseSubtypeValues(src string) (map[string]struct{}, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "subtypes.go", src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	values := make(map[string]struct{})
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.CONST {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			if !isSubtypeTypeRef(vs.Type) {
				continue
			}
			for _, v := range vs.Values {
				lit, ok := v.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				values[unquoteSimple(lit.Value)] = struct{}{}
			}
		}
	}
	return values, nil
}

func isSubtypeTypeRef(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name == "Subtype"
	case *ast.SelectorExpr:
		return t.Sel != nil && t.Sel.Name == "Subtype"
	}
	return false
}
