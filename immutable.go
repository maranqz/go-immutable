package immutable

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

var ImmutableAnalyzer = &analysis.Analyzer{
	Name:     "immutable",
	Doc:      "finds attempts to change read-only values",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {

		v := &visiter{readonlyObjects: map[*ast.Object]*object{}}
		v.walk(file)

		for _, obj := range v.tryMutate {
			pass.Report(analysis.Diagnostic{
				Pos:     obj.ident.Pos(),
				Message: fmt.Sprintf("try to change %s", obj.ident.Name),
			})
		}
	}

	return nil, nil
}

type visiter struct {
	readonlyObjects map[*ast.Object]*object
	tryMutate       []*object
}

type object struct {
	ident     *ast.Ident
	root      *ast.Object
	isPointer bool
}

func (v *visiter) walk(n ast.Node) {
	if n != nil {
		ast.Walk(v, n)
	}
}

func (v *visiter) Visit(n ast.Node) ast.Visitor {
	inc, ok := n.(*ast.IncDecStmt)
	if ok {
		l, ok := inc.X.(*ast.StarExpr)
		if ok {
			n = l.X
		}

		lId, ok := inc.X.(*ast.Ident)
		if !ok {
			return v
		}

		if v.isReadOnly(lId.Obj) {
			ro := v.getReadOnly(lId.Obj)

			v.tryMutate = append(v.tryMutate, &object{
				ident:     lId,
				root:      ro.root,
				isPointer: ro.isPointer,
			})
		}

		return v
	}

	assign, ok := n.(*ast.AssignStmt)
	if !ok {
		return v
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

	if v.isReadOnly(lId.Obj) {
		ro := v.getReadOnly(lId.Obj)
		if isPtrAssign {
			if !ro.isPointer {
				return v
			}
		}

		v.tryMutate = append(v.tryMutate, &object{
			ident:     lId,
			root:      ro.root,
			isPointer: ro.isPointer,
		})
	}

	// TODO replace on more useful configuration
	if strings.HasSuffix(lId.Name, "RO") {
		v.readonlyObjects[lId.Obj] = &object{
			ident: lId,
		}
	}

	// read Rhs
	rhs := assign.Rhs[0]
	if rId, ok := rhs.(*ast.Ident); ok {
		if v.isReadOnly(rId.Obj) &&
			v.getReadOnly(rId.Obj).isPointer {
			v.readonlyObjects[lId.Obj] = &object{
				ident:     lId,
				root:      rId.Obj,
				isPointer: true,
			}
		}

		return v
	}

	if rId, ok := rhs.(*ast.UnaryExpr); ok && rId.Op == token.AND {
		rEl, ok := rId.X.(*ast.Ident)
		if ok {
			if v.isReadOnly(rEl.Obj) {
				v.readonlyObjects[lId.Obj] = &object{
					ident:     lId,
					root:      rEl.Obj,
					isPointer: true,
				}
			}
		}
	}

	return v
}

func (v *visiter) isReadOnly(o *ast.Object) bool {
	_, ok := v.readonlyObjects[o]

	return ok
}

func (v *visiter) getReadOnly(o *ast.Object) *object {
	return v.readonlyObjects[o]
}
