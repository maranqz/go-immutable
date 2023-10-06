package immutable

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
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

var analyse = globalAnalyse{
	readonlyObjects: map[types.Object]*object{},
}

func run(pass *analysis.Pass) (interface{}, error) {
	// TODO do concurrency processing
	fmt.Println(pass.Pkg.Path())

	for _, file := range pass.Files {
		v := &visiter{
			globalAnalyse: analyse,
			pass:          pass,
		}
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

type globalAnalyse struct {
	readonlyObjects map[types.Object]*object
	tryMutate       []*object
}

type visiter struct {
	globalAnalyse
	pass *analysis.Pass
}

type object struct {
	ident     *ast.Ident
	obj       types.Object
	parent    types.Object
	root      types.Object
	isPointer bool
}

func (v *visiter) ObjUnar(ident *ast.Ident) *object {
	obj := v.pass.TypesInfo.ObjectOf(ident)
	p := v.getReadOnly(obj)

	if p.root != nil {
		p = v.getReadOnly(p.root)
	}

	return &object{
		ident:     ident,
		obj:       obj,
		root:      p.obj,
		isPointer: p.isPointer,
	}
}

func (v *visiter) Obj(ident, pIdent *ast.Ident) *object {
	return v.obj(ident, pIdent, false)
}

func (v *visiter) ObjPtr(ident, pIdent *ast.Ident) *object {
	return v.obj(ident, pIdent, true)
}

func (v *visiter) obj(ident, pIdent *ast.Ident, isPtr bool) *object {
	pObj := v.pass.TypesInfo.ObjectOf(pIdent)
	p := v.getReadOnly(pObj)
	root := p

	if root.root != nil {
		root = v.getReadOnly(root.root)
	}

	obj := v.pass.TypesInfo.ObjectOf(ident)

	return &object{
		ident:     ident,
		obj:       obj,
		parent:    p.obj,
		root:      root.obj,
		isPointer: isPtr || p.isPointer,
	}
}

func (v *visiter) walk(n ast.Node) {
	if n != nil {
		ast.Walk(v, n)
	}
}

func (v *visiter) Visit(n ast.Node) ast.Visitor {
	inc, ok := n.(*ast.IncDecStmt)
	if ok {
		x := inc.X
		l, ok := x.(*ast.StarExpr)
		if ok {
			x = l.X
		}

		sX, ok := x.(*ast.SelectorExpr)
		if ok {
			x = sX.Sel
		}

		lId, ok := x.(*ast.Ident)
		if !ok {
			return v
		}

		lObj := v.pass.TypesInfo.ObjectOf(lId)
		if v.isReadOnly(lObj) {
			v.tryMutate = append(v.tryMutate, v.ObjUnar(lId))
		}

		return v
	}

	valueSpec, ok := n.(*ast.ValueSpec)
	if ok {
		// TODO multiple assign
		if len(valueSpec.Names) == 0 {
			return v
		}

		lId := valueSpec.Names[0]
		lObj := v.pass.TypesInfo.ObjectOf(lId)
		v.markROByName(lId)

		if rId, ok := v.isPtrRight(valueSpec.Values[0]); ok {
			rObj := v.pass.TypesInfo.ObjectOf(rId)
			if v.isReadOnly(rObj) {
				v.readonlyObjects[lObj] = v.ObjPtr(lId, rId)
			}
		}

		return v
	}

	assign, ok := n.(*ast.AssignStmt)
	if !ok {
		return v
	}

	// TODO multiple assign
	isPtrAssign := false
	lId, ok := assign.Lhs[0].(*ast.Ident)
	if !ok {
		start, ok := assign.Lhs[0].(*ast.StarExpr)
		if ok {
			lId = start.X.(*ast.Ident)
			isPtrAssign = ok
		}
	}

	lObj := v.pass.TypesInfo.ObjectOf(lId)
	if v.isReadOnly(lObj) {
		ro := v.getReadOnly(lObj)
		if isPtrAssign {
			if !ro.isPointer {
				return v
			}
		}

		v.tryMutate = append(v.tryMutate, v.Obj(lId, ro.ident))
	}

	// TODO replace on more useful configuration
	v.markROByName(lId)

	// read Rhs
	rhs := assign.Rhs[0]
	if rId, ok := rhs.(*ast.Ident); ok {
		v.markROByRight(lId, rId)

		return v
	}

	if rId, ok := v.isPtrRight(rhs); ok {
		rObj := v.pass.TypesInfo.ObjectOf(rId)
		if v.isReadOnly(rObj) {
			v.readonlyObjects[lObj] = v.ObjPtr(lId, rId)
		}
	}

	return v
}

func (v *visiter) isReadOnly(o types.Object) bool {
	_, ok := v.readonlyObjects[o]

	return ok
}

func (v *visiter) getReadOnly(o types.Object) *object {
	return v.readonlyObjects[o]
}

func (v *visiter) markROByName(ident *ast.Ident) {
	if strings.HasSuffix(ident.Name, "RO") {
		obj := v.pass.TypesInfo.ObjectOf(ident)
		v.readonlyObjects[obj] = &object{
			ident: ident,
			obj:   obj,
		}
	}
}

func (v *visiter) markROByRight(lId, rId *ast.Ident) {
	lObj := v.pass.TypesInfo.ObjectOf(lId)
	rObj := v.pass.TypesInfo.ObjectOf(rId)

	if v.isReadOnly(rObj) &&
		v.getReadOnly(rObj).isPointer {
		v.readonlyObjects[lObj] = v.Obj(lId, rId)
	}
}

func (v *visiter) isPtrRight(n ast.Node) (*ast.Ident, bool) {
	if rId, ok := n.(*ast.UnaryExpr); ok && rId.Op == token.AND {
		x := rId.X
		if sx, ok := x.(*ast.SelectorExpr); ok {
			x = sx.Sel
		}

		rEl, ok := x.(*ast.Ident)

		return rEl, ok
	}

	return nil, false
}
