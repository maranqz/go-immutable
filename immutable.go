package immutable

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var ImmutableAnalyzer = &analysis.Analyzer{
	Name: "immutable",
	Doc:  "finds attempts to change read-only values",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			// TODO
			// https://github.com/sourcegraph/go-template-lint/tree/master

			return true
		})
	}

	return nil, nil
}
