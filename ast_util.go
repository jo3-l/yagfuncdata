package yagfuncdata

import (
	"go/ast"
	"go/token"
	"strconv"
)

func lookupFuncDecl(f *ast.File, name string) (*ast.FuncDecl, bool) {
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == name {
			return fn, true
		}
	}
	return nil, false
}

func lookupVarDecl(f *ast.File, name string) (initExpr ast.Expr, ok bool) {
	for _, decl := range f.Decls {
		if decl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range decl.Specs {
				if spec, ok := spec.(*ast.ValueSpec); ok {
					for i := range spec.Names {
						if spec.Names[i].Name == name {
							return spec.Values[i], true
						}
					}
				}
			}
		}
	}
	return nil, false
}

func unpackCompositeLiteral(expr ast.Expr) (entries []*ast.KeyValueExpr, ok bool) {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil, false
	}

	entries = make([]*ast.KeyValueExpr, 0, len(lit.Elts))
	for _, elt := range lit.Elts {
		if entry, ok := elt.(*ast.KeyValueExpr); ok {
			entries = append(entries, entry)
		}
	}
	return entries, true
}

func unpackStringLit(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}

	inner, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return inner, true
}
