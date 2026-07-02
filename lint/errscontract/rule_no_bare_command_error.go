// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package errscontract

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// CheckNoBareCommandError rejects bare fmt.Errorf / errors.New calls used
// as the final error returned from command boundaries (cobra command RunE /
// Run functions and shortcut Validate / Execute functions). Intermediate
// wrapping for logging is fine, but the value that crosses the command
// boundary must be a typed errs.* error so the dispatcher can emit a
// machine-readable JSON envelope.
func CheckNoBareCommandError(path, src string) []Violation {
	path = filepathToSlash(path)
	if !isCommandBoundaryScope(path) {
		return nil
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, 0)
	if err != nil {
		return nil
	}
	boundaries := buildBoundaryIndex(file, fset, path)
	var out []Violation
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			out = append(out, collectBareCommandErrorReturns(path, fset, node.Body, boundaries)...)
		case *ast.FuncLit:
			out = append(out, collectBareCommandErrorReturns(path, fset, node.Body, boundaries)...)
		}
		return true
	})
	return out
}

func collectBareCommandErrorReturns(path string, fset *token.FileSet, body *ast.BlockStmt, boundaries boundaryIndex) []Violation {
	if body == nil {
		return nil
	}
	var out []Violation
	seen := map[int]bool{}
	scanBlock(path, fset, body, boundaries, seen, &out)
	return out
}

func scanBlock(path string, fset *token.FileSet, body *ast.BlockStmt, boundaries boundaryIndex, seen map[int]bool, out *[]Violation) {
	if body == nil {
		return
	}
	for _, stmt := range body.List {
		scanStmt(path, fset, stmt, boundaries, seen, out)
	}
}

func scanStmt(path string, fset *token.FileSet, stmt ast.Stmt, boundaries boundaryIndex, seen map[int]bool, out *[]Violation) {
	switch node := stmt.(type) {
	case *ast.ReturnStmt:
		line := fset.Position(node.Pos()).Line
		if !boundaries.ContainsReturn(path, line) {
			return
		}
		for _, result := range node.Results {
			if call := bareErrorCall(result); call != nil {
				appendBareCommandErrorViolation(path, fset, call, seen, out)
			}
		}
	case *ast.BlockStmt:
		scanBlock(path, fset, node, boundaries, seen, out)
	case *ast.IfStmt:
		if node.Init != nil {
			scanStmt(path, fset, node.Init, boundaries, seen, out)
		}
		scanBlock(path, fset, node.Body, boundaries, seen, out)
		if node.Else != nil {
			scanStmt(path, fset, node.Else, boundaries, seen, out)
		}
	case *ast.ForStmt:
		if node.Init != nil {
			scanStmt(path, fset, node.Init, boundaries, seen, out)
		}
		scanBlock(path, fset, node.Body, boundaries, seen, out)
	case *ast.RangeStmt:
		scanBlock(path, fset, node.Body, boundaries, seen, out)
	case *ast.SwitchStmt:
		if node.Init != nil {
			scanStmt(path, fset, node.Init, boundaries, seen, out)
		}
		for _, stmt := range node.Body.List {
			if clause, ok := stmt.(*ast.CaseClause); ok {
				scanStmtList(path, fset, clause.Body, boundaries, seen, out)
			}
		}
	case *ast.TypeSwitchStmt:
		if node.Init != nil {
			scanStmt(path, fset, node.Init, boundaries, seen, out)
		}
		for _, stmt := range node.Body.List {
			if clause, ok := stmt.(*ast.CaseClause); ok {
				scanStmtList(path, fset, clause.Body, boundaries, seen, out)
			}
		}
	case *ast.SelectStmt:
		for _, stmt := range node.Body.List {
			if clause, ok := stmt.(*ast.CommClause); ok {
				scanStmtList(path, fset, clause.Body, boundaries, seen, out)
			}
		}
	}
}

func scanStmtList(path string, fset *token.FileSet, stmts []ast.Stmt, boundaries boundaryIndex, seen map[int]bool, out *[]Violation) {
	for _, stmt := range stmts {
		scanStmt(path, fset, stmt, boundaries, seen, out)
	}
}

func appendBareCommandErrorViolation(path string, fset *token.FileSet, call *ast.CallExpr, seen map[int]bool, out *[]Violation) {
	pos := fset.Position(call.Pos())
	if seen[pos.Line] {
		return
	}
	seen[pos.Line] = true
	*out = append(*out, Violation{
		Rule:       "no_bare_command_error",
		Action:     ActionReject,
		File:       path,
		Line:       pos.Line,
		Message:    "command boundary errors must use typed structured errors",
		Suggestion: "return typed errs.* errors with param/hint metadata so callers receive machine-readable error JSON",
	})
}

func bareErrorCall(expr ast.Expr) *ast.CallExpr {
	switch v := expr.(type) {
	case *ast.ParenExpr:
		return bareErrorCall(v.X)
	case *ast.CallExpr:
		if isBareCommandErrorCall(selectorName(v.Fun)) {
			return v
		}
	}
	return nil
}

func isBareCommandErrorCall(name string) bool {
	return name == "fmt.Errorf" || name == "errors.New"
}

func selectorName(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		prefix := selectorName(v.X)
		if prefix == "" {
			return v.Sel.Name
		}
		return prefix + "." + v.Sel.Name
	default:
		return ""
	}
}

func isCommandBoundaryScope(path string) bool {
	path = filepathToSlash(path)
	return (strings.HasPrefix(path, "cmd/") || strings.HasPrefix(path, "shortcuts/")) &&
		strings.HasSuffix(path, ".go") &&
		!strings.HasSuffix(path, "_test.go")
}

type fileLine struct {
	file string
	line int
}

type boundaryIndex struct {
	Returns map[fileLine]bool
	Funcs   map[string]bool
}

func (idx boundaryIndex) ContainsReturn(path string, line int) bool {
	if idx.Returns == nil {
		return false
	}
	return idx.Returns[fileLine{file: filepathToSlash(path), line: line}]
}

func buildBoundaryIndex(file *ast.File, fset *token.FileSet, path string) boundaryIndex {
	idx := boundaryIndex{
		Returns: map[fileLine]bool{},
		Funcs:   map[string]bool{},
	}
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		switch {
		case isCobraCommandLiteral(lit):
			markBoundaryFields(idx, fset, path, lit, "RunE", "Run")
		case isShortcutLiteral(lit):
			markBoundaryFields(idx, fset, path, lit, "Validate", "Execute")
		}
		return true
	})
	markBoundaryAssignments(file, fset, path, idx)
	markBoundaryFunctionReturns(file, fset, path, idx)
	return idx
}

func markBoundaryFields(idx boundaryIndex, fset *token.FileSet, path string, lit *ast.CompositeLit, names ...string) {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok || !isBoundaryField(kv.Key, names...) {
			continue
		}
		markBoundaryExpr(idx, fset, path, kv.Value)
	}
}

func markBoundaryAssignments(file *ast.File, fset *token.FileSet, path string, idx boundaryIndex) {
	ast.Inspect(file, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for i, lhs := range assign.Lhs {
			sel, ok := lhs.(*ast.SelectorExpr)
			if !ok || !isBoundaryAssignmentField(path, sel.Sel.Name) {
				continue
			}
			var rhs ast.Expr
			if len(assign.Rhs) == 1 {
				rhs = assign.Rhs[0]
			} else if i < len(assign.Rhs) {
				rhs = assign.Rhs[i]
			}
			if rhs != nil {
				markBoundaryExpr(idx, fset, path, rhs)
			}
		}
		return true
	})
}

func markBoundaryExpr(idx boundaryIndex, fset *token.FileSet, path string, expr ast.Expr) {
	switch v := expr.(type) {
	case *ast.FuncLit:
		markReturnStatements(idx, fset, path, v.Body)
	case *ast.Ident:
		idx.Funcs[v.Name] = true
	case *ast.SelectorExpr:
		idx.Funcs[v.Sel.Name] = true
	}
}

func markBoundaryFunctionReturns(file *ast.File, fset *token.FileSet, path string, idx boundaryIndex) {
	if len(idx.Funcs) == 0 {
		return
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Body == nil || !idx.Funcs[fn.Name.Name] {
			continue
		}
		markReturnStatements(idx, fset, path, fn.Body)
	}
}

func markReturnStatements(idx boundaryIndex, fset *token.FileSet, path string, body *ast.BlockStmt) {
	ast.Inspect(body, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		line := fset.Position(ret.Pos()).Line
		idx.Returns[fileLine{file: filepathToSlash(path), line: line}] = true
		return true
	})
}

func isCobraCommandLiteral(lit *ast.CompositeLit) bool {
	return commandTypeName(lit.Type) == "cobra.Command" || commandTypeName(lit.Type) == "Command"
}

func isShortcutLiteral(lit *ast.CompositeLit) bool {
	return commandTypeName(lit.Type) == "common.Shortcut" || commandTypeName(lit.Type) == "Shortcut"
}

func commandTypeName(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		prefix := commandTypeName(v.X)
		if prefix == "" {
			return v.Sel.Name
		}
		return prefix + "." + v.Sel.Name
	}
	return ""
}

func isBoundaryField(expr ast.Expr, names ...string) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	for _, name := range names {
		if ident.Name == name {
			return true
		}
	}
	return false
}

func isBoundaryAssignmentField(path, name string) bool {
	path = filepathToSlash(path)
	switch {
	case strings.HasPrefix(path, "cmd/"):
		return name == "RunE" || name == "Run"
	case strings.HasPrefix(path, "shortcuts/"):
		return name == "Validate" || name == "Execute"
	default:
		return false
	}
}

func filepathToSlash(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
