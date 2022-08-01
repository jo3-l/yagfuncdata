package yagfuncdata

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// A Source is a source of YAGPDB template function information.
type Source interface {
	Fetch(ctx context.Context) (funcs []string, err error)
}

// DefaultSources returns a set of builtin sources that use the given
// FileContentProvider.
func DefaultSources(fcp FileContentProvider) []Source {
	return []Source{
		NewBaseContextFuncSource(fcp),
		NewBuiltinFuncSource(fcp),
		NewLogsPluginExtensionFuncSource(fcp),
		NewTicketsPluginExtensionFuncSource(fcp),
		NewCommandsPluginExtensionFuncSource(fcp),
		NewCustomCommandsPluginExtensionFuncSource(fcp),
	}
}

var _ Source = (*BaseContextFuncSource)(nil)

// NewBaseContextFuncSource creates a new Source that provides information
// regarding functions defined in common/templates/context.go, which include
// base context and standard functions.
func NewBaseContextFuncSource(fcp FileContentProvider) *BaseContextFuncSource {
	return &BaseContextFuncSource{fcp}
}

type BaseContextFuncSource struct {
	fcp FileContentProvider
}

func (c *BaseContextFuncSource) Fetch(ctx context.Context) ([]string, error) {
	const filepath = "common/templates/context.go"
	src, err := c.fcp.Get(ctx, filepath)
	if err != nil {
		return nil, fmt.Errorf("fetching base context functions: %w", err)
	}

	f, err := parser.ParseFile(token.NewFileSet(), filepath, src, parser.Mode(0))
	if err != nil {
		return nil, fmt.Errorf("fetching base context functions: %s contains invalid Go code: %w", filepath, err)
	}

	contextFuncs, err := c.collectStandardFuncs(f)
	if err != nil {
		return nil, fmt.Errorf("fetching base context functions: %w", err)
	}
	standardFuncs, err := c.collectStandardContextFuncs(f)
	if err != nil {
		return nil, fmt.Errorf("fetching base context functions: %w", err)
	}
	return append(contextFuncs, standardFuncs...), nil
}

func (c *BaseContextFuncSource) collectStandardFuncs(f *ast.File) ([]string, error) {
	// [...]
	// func baseContextFuncs(c *Context) {
	// 	 c.addContextFunc(name1, func1)
	//   c.addContextFunc(name2, func2)
	//   [...]
	// }
	// [...]
	fn, ok := lookupFuncDecl(f, "baseContextFuncs")
	if !ok {
		return nil, errors.New("no definition for baseContextFuncs")
	}

	var funcs []string
	for _, stmt := range fn.Body.List {
		if expr, ok := stmt.(*ast.ExprStmt); ok {
			if call, ok := expr.X.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "addContextFunc" {
					if len(call.Args) > 0 {
						if name, ok := unpackStringLit(call.Args[0]); ok {
							funcs = append(funcs, name)
						}
					}
				}
			}
		}
	}
	return funcs, nil
}

func (c *BaseContextFuncSource) collectStandardContextFuncs(f *ast.File) ([]string, error) {
	// var (
	//   StandardFuncMap = map[string]interface{}{
	//     name1: func1,
	//     name2: func2,
	//     [...]
	//   }
	// )
	initExpr, ok := lookupVarDecl(f, "StandardFuncMap")
	if !ok {
		return nil, errors.New("no definition for StandardFuncMap")
	}

	entries, ok := unpackCompositeLiteral(initExpr)
	if !ok {
		return nil, errors.New("initializer for StandardFuncMap is not a composite literal")
	}

	var funcs []string
	for _, entry := range entries {
		if name, ok := unpackStringLit(entry.Key); ok {
			funcs = append(funcs, name)
		}
	}
	return funcs, nil
}

var _ Source = (*BuiltinFuncSource)(nil)

// NewBuiltinFuncSource creates a new Source that provides information regarding
// builtin template functions defined in lib/template/funcs.go.
func NewBuiltinFuncSource(fcp FileContentProvider) *BuiltinFuncSource {
	return &BuiltinFuncSource{fcp}
}

type BuiltinFuncSource struct {
	fcp FileContentProvider
}

func (b *BuiltinFuncSource) Fetch(ctx context.Context) ([]string, error) {
	const filepath = "lib/template/funcs.go"
	src, err := b.fcp.Get(ctx, filepath)
	if err != nil {
		return nil, fmt.Errorf("fetching builtin functions: %w", err)
	}

	f, err := parser.ParseFile(token.NewFileSet(), filepath, src, parser.Mode(0))
	if err != nil {
		return nil, fmt.Errorf("fetching builtin functions: %s contains invalid Go code: %w", filepath, err)
	}

	// func builtins() FuncMap {
	//   return FuncMap{
	//     name1: func1,
	//     name2: func2,
	//     [...]
	//   }
	// }
	fn, ok := lookupFuncDecl(f, "builtins")
	if !ok {
		return nil, errors.New("fetching builtin functions: no definition for builtins")
	}

	if len(fn.Body.List) == 0 {
		return nil, errors.New("fetching builtin functions: builtins has empty body")
	}
	ret, ok := fn.Body.List[0].(*ast.ReturnStmt)
	if !ok {
		return nil, errors.New("fetching builtin functions: no return statement in builtins")
	}

	if len(ret.Results) == 0 {
		return nil, errors.New("fetching builtin functions: return statement in builtins has no results")
	}
	entries, ok := unpackCompositeLiteral(ret.Results[0])
	if !ok {
		return nil, errors.New("fetching builtin functions: result of return statement in builtins is not a composite literal")
	}

	var funcs []string
	for _, entry := range entries {
		if name, ok := unpackStringLit(entry.Key); ok {
			funcs = append(funcs, name)
		}
	}
	return funcs, nil
}

// NewLogsPluginExtensionFuncSource creates a new Source that provides
// information regarding extension template functions registered by the logs
// plugin in logs/template_extensions.go.
func NewLogsPluginExtensionFuncSource(fcp FileContentProvider) *PluginExtensionFuncSource {
	return &PluginExtensionFuncSource{fcp, "logs/template_extensions.go"}
}

// NewTicketsPluginExtensionFuncSource creates a new Source that provides
// information regarding extension template functions registered by the tickets
// plugin in tickets/tmplextensions.go.
func NewTicketsPluginExtensionFuncSource(fcp FileContentProvider) *PluginExtensionFuncSource {
	return &PluginExtensionFuncSource{fcp, "tickets/tmplextensions.go"}
}

// NewCommandsPluginExtensionFuncSource creates a new Source that provides
// information regarding extension template functions registered by the command
// plugin in commands/tmplexec.go.
func NewCommandsPluginExtensionFuncSource(fcp FileContentProvider) *PluginExtensionFuncSource {
	return &PluginExtensionFuncSource{fcp, "commands/tmplexec.go"}
}

// NewCustomCommandsPluginExtensionFuncSource creates a new Source that provides
// information regarding extension template functions registered by the custom commands
// plugin in customcommands/tmplextensions.go
func NewCustomCommandsPluginExtensionFuncSource(fcp FileContentProvider) *PluginExtensionFuncSource {
	return &PluginExtensionFuncSource{fcp, "customcommands/tmplextensions.go"}
}

var _ Source = (*PluginExtensionFuncSource)(nil)

// A PluginExtensionFuncSource provides information regarding template functions
// added by a plugin. For example, the logs plugin registers pastUsernames and
// pastNicknames in logs/template_extensions.go.
type PluginExtensionFuncSource struct {
	fcp      FileContentProvider
	filepath string
}

func (p *PluginExtensionFuncSource) Fetch(ctx context.Context) ([]string, error) {
	src, err := p.fcp.Get(ctx, p.filepath)
	if err != nil {
		return nil, fmt.Errorf("fetching plugin extension functions: %w", err)
	}

	f, err := parser.ParseFile(token.NewFileSet(), p.filepath, src, parser.Mode(0))
	if err != nil {
		return nil, fmt.Errorf("fetching plugin extension functions: %s contains invalid Go code: %w", p.filepath, err)
	}

	// [...]
	// templates.RegisterSetupFunc(func(ctx *templates.Context) {
	//   ctx.ContextFuncs[name1] = func1
	//   ctx.ContextFuncs[name2] = func2
	//   ...
	// })
	// [...]
	var funcs []string
	ast.Inspect(f, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Only consider calls to [...].RegisterSetupFunc with precisely one argument.
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if sel.Sel.Name != "RegisterSetupFunc" || len(call.Args) != 1 {
			return true
		}

		fn, ok := call.Args[0].(*ast.FuncLit)
		if !ok {
			return true
		}

		for _, stmt := range fn.Body.List {
			// Only consider assignments like [...].ContextFuncs[name] = ...
			if assign, ok := stmt.(*ast.AssignStmt); ok && len(assign.Lhs) == 1 {
				if index, ok := assign.Lhs[0].(*ast.IndexExpr); ok {
					if sel, ok := index.X.(*ast.SelectorExpr); ok && sel.Sel.Name == "ContextFuncs" {
						if name, ok := unpackStringLit(index.Index); ok {
							funcs = append(funcs, name)
						}
					}
				}
			}
		}
		return true
	})
	return funcs, nil
}
