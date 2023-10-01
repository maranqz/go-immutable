package immutable

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var ImmutableAnalyzer = &analysis.Analyzer{
	Name:     "immutable",
	Doc:      "finds attempts to change read-only values",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {

		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

		// the inspector has a `filter` feature that enables type-based filtering
		// The anonymous function will be only called for the ast nodes whose type
		// matches an element in the filter
		nodeFilter := []ast.Node{
			(*ast.AssignStmt)(nil),
		}

		// TODO combine in struct, add first initialization, name
		readOnlyObjs := map[*ast.Object]struct{}{}
		readOnlyPtrObjs := map[*ast.Object]struct{}{}

		// TODO migrate to ast.Walk
		inspect.Preorder(nodeFilter, func(n ast.Node) {
			assign, ok := n.(*ast.AssignStmt)
			if !ok {
				return
			}

			// TODO multiple assign
			// read Lhs
			isPtrAssign := false
			lId, ok := assign.Lhs[0].(*ast.Ident)
			if !ok {
				start, ok := assign.Lhs[0].(*ast.StarExpr)
				isPtrAssign = ok
				if ok {
					lId = start.X.(*ast.Ident)
				}
			}

			if _, isRO := readOnlyObjs[lId.Obj]; isRO {
				if isPtrAssign {
					_, isPtr := readOnlyPtrObjs[lId.Obj]
					if !isPtr {
						return
					}
				}

				pass.Report(analysis.Diagnostic{
					Pos: lId.Pos(),
					// TODO use descriptive message
					Message: fmt.Sprintf("try to change %s", lId.Name),
				})
			}

			// TODO replace on more useful configuration
			if strings.HasSuffix(lId.Name, "RO") {
				readOnlyObjs[lId.Obj] = struct{}{}
			}

			// read Rhs
			rhs := assign.Rhs[0]
			if rId, ok := rhs.(*ast.Ident); ok {
				if _, isRO := readOnlyPtrObjs[rId.Obj]; isRO {
					readOnlyObjs[lId.Obj] = struct{}{}
					readOnlyPtrObjs[lId.Obj] = struct{}{}
				}

				return
			}

			if rId, ok := rhs.(*ast.UnaryExpr); ok && rId.Op == token.AND {
				rEl, ok := rId.X.(*ast.Ident)
				if ok {
					if _, isRO := readOnlyObjs[rEl.Obj]; isRO {
						readOnlyObjs[lId.Obj] = struct{}{}
						readOnlyPtrObjs[lId.Obj] = struct{}{}
					}
				}
			}

			return
		})

		ast.Inspect(file, func(n ast.Node) bool {
			// TODO
			// https://github.com/sourcegraph/go-template-lint/tree/master

			return true
		})
	}

	return nil, nil
}
