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

//nolint:gochecknoglobals // linter configuration for Analyzer
var ImmutableAnalyzer = &analysis.Analyzer{
	Name:     "immutable",
	Doc:      "finds attempts to change read-only values",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

//nolint:gochecknoglobals // use to store data from all packages
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

	//nolint:nilnil
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

func (o *object) isPtr() bool {
	return o != nil && o.isPointer
}

func (v *visiter) ObjUnar(ident *ast.Ident) *object {
	obj := v.pass.TypesInfo.ObjectOf(ident)
	parent := v.getReadOnly(obj)

	var parentIdent *ast.Ident
	if parent != nil {
		parentIdent = parent.ident
	}

	return v.obj(ident, parentIdent, false)
}

func (v *visiter) Obj(ident, pIdent *ast.Ident) *object {
	return v.obj(ident, pIdent, false)
}

func (v *visiter) ObjPtr(ident, pIdent *ast.Ident) *object {
	return v.obj(ident, pIdent, true)
}

func (v *visiter) obj(ident, pIdent *ast.Ident, isPtr bool) *object {
	parentObj := v.pass.TypesInfo.ObjectOf(pIdent)
	parent := v.getReadOnly(parentObj)
	rootObj := parentObj

	if parent != nil &&
		parent.root != nil &&
		v.isReadOnly(parent.root) {
		rootObj = v.getReadOnly(parent.root).obj
	}

	obj := v.pass.TypesInfo.ObjectOf(ident)

	return &object{
		ident:     ident,
		obj:       obj,
		parent:    parentObj,
		root:      rootObj,
		isPointer: isPtr || parent.isPtr(),
	}
}

func (v *visiter) walk(n ast.Node) {
	if n != nil {
		ast.Walk(v, n)
	}
}

//nolint:nestif,funlen,gocognit,cyclop // FIXME
func (v *visiter) Visit(node ast.Node) ast.Visitor {
	// Mark struct
	typeSpec, ok := node.(*ast.TypeSpec)
	if ok {
		v.markROByName(typeSpec.Name)

		return v
	}

	inc, ok := node.(*ast.IncDecStmt)
	if ok {
		expr := inc.X

		ptr, ok := expr.(*ast.StarExpr)
		if ok {
			expr = ptr.X
		}

		selectorExpr, ok := expr.(*ast.SelectorExpr)
		if ok {
			if ident, structNamed, ok := v.getStructNamed(selectorExpr); ok {
				obj := structNamed.Obj()
				if v.isReadOnly(obj) {
					ro := v.getReadOnly(obj)
					v.tryMutate = append(v.tryMutate, v.Obj(ident, ro.ident))
				}

				return v
			}

			expr = selectorExpr.Sel
		}

		lIdent, ok := expr.(*ast.Ident)
		if !ok {
			return v
		}

		lObj := v.pass.TypesInfo.ObjectOf(lIdent)
		if v.isReadOnly(lObj) {
			v.tryMutate = append(v.tryMutate, v.ObjUnar(lIdent))
		}

		return v
	}

	valueSpec, ok := node.(*ast.ValueSpec)
	if ok {
		// TODO multiple assign
		if len(valueSpec.Names) == 0 {
			return v
		}

		lIdent := valueSpec.Names[0]
		lObj := v.pass.TypesInfo.ObjectOf(lIdent)
		v.markROByName(lIdent)

		if rIdent, ok := v.isPtrRight(valueSpec.Values[0]); ok {
			rObj := v.pass.TypesInfo.ObjectOf(rIdent)
			if v.isReadOnly(rObj) {
				v.readonlyObjects[lObj] = v.ObjPtr(lIdent, rIdent)
			}
		}

		return v
	}

	assign, ok := node.(*ast.AssignStmt)
	if !ok {
		return v
	}

	// TODO multiple assign
	var (
		lIdent *ast.Ident
		lObj   types.Object
	)

	lhsFirst := assign.Lhs[0]

	star, isPtrAssign := lhsFirst.(*ast.StarExpr)
	if isPtrAssign {
		lhsFirst = star.X
	}

	// struct
	lSelectorExpr, ok := lhsFirst.(*ast.SelectorExpr)
	if ok {
		if ident, structNamed, ok := v.getStructNamed(lSelectorExpr); ok {
			obj := structNamed.Obj()
			if v.isReadOnly(obj) {
				ro := v.getReadOnly(obj)
				v.tryMutate = append(v.tryMutate, v.Obj(ident, ro.ident))
			}

			lIdent = ident
			lObj = v.pass.TypesInfo.ObjectOf(lIdent)
		} else {
			v.unexpected(lSelectorExpr)

			return v
		}
	} else {
		// variable
		lIdent, ok = lhsFirst.(*ast.Ident)
		if !ok {
			v.unexpected(lSelectorExpr)

			return v
		}

		lObj = v.pass.TypesInfo.ObjectOf(lIdent)
		if v.isReadOnly(lObj) {
			ro := v.getReadOnly(lObj)
			if isPtrAssign && !ro.isPtr() {
				return v
			}

			v.tryMutate = append(v.tryMutate, v.Obj(lIdent, ro.ident))
		}
	}

	// TODO replace on more useful configuration
	// Mark variable
	v.markROByName(lIdent)

	// read Rhs
	rhs := assign.Rhs[0]
	if rIdent, ok := rhs.(*ast.Ident); ok {
		v.markROByRight(lIdent, rIdent)

		return v
	}

	if rIdent, ok := v.isPtrRight(rhs); ok {
		rObj := v.pass.TypesInfo.ObjectOf(rIdent)
		if v.isReadOnly(rObj) {
			v.readonlyObjects[lObj] = v.ObjPtr(lIdent, rIdent)
		}
	}

	return v
}

func (v *visiter) getStructNamed(selectorExpr *ast.SelectorExpr) (
	*ast.Ident,
	*types.Named,
	bool,
) {
	ident, ok := selectorExpr.X.(*ast.Ident)
	if ok {
		obj := v.pass.TypesInfo.ObjectOf(ident)
		objType := obj.Type()

		if typePtr, ok := objType.(*types.Pointer); ok {
			objType = typePtr.Elem()
		}

		if typeNamed, ok := objType.(*types.Named); ok {
			return ident, typeNamed, true
		}
	}

	return nil, nil, false
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
		v.readonlyObjects[obj] = v.ObjUnar(ident)
	}
}

func (v *visiter) markROByRight(lIdent, rIdent *ast.Ident) {
	lObj := v.pass.TypesInfo.ObjectOf(lIdent)
	rObj := v.pass.TypesInfo.ObjectOf(rIdent)

	if v.isReadOnly(rObj) &&
		v.getReadOnly(rObj).isPtr() {
		v.readonlyObjects[lObj] = v.Obj(lIdent, rIdent)
	}
}

func (v *visiter) isPtrRight(n ast.Node) (*ast.Ident, bool) {
	if unaryExpr, ok := n.(*ast.UnaryExpr); ok &&
		unaryExpr.Op == token.AND {
		expr := unaryExpr.X
		if selectorExpr, ok := expr.(*ast.SelectorExpr); ok {
			expr = selectorExpr.Sel
		}

		ident, ok := expr.(*ast.Ident)

		return ident, ok
	}

	return nil, false
}

func (v *visiter) unexpected(n ast.Node) {
	v.pass.Reportf(
		n.Pos(),
		`unexpected code, use "// nolint:immutable" and create issue about that with example`,
	)
}
