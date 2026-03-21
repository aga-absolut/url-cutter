package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer — анализатор, запрещающий os.Exit в func main() пакета main.
var Analyzer = &analysis.Analyzer{
	Name: "noexit",
	Doc:  `Запрещает прямой вызов os.Exit в функции main`,
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			pkg, ok := sel.X.(*ast.Ident)
			if !ok || pkg.Name != "os" {
				return true
			}

			if sel.Sel.Name == "Exit" {
				pass.Reportf(call.Pos(), "вызов os.Exit в main запрещён")
			}

			return true
		})
	}

	return nil, nil
}
