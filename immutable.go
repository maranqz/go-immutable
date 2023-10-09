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
	ident         *ast.Ident
	obj           types.Object
	parent        types.Object
	root          types.Object
	rootVar       types.Object
	isPointer     bool
	isForExported bool
}

func (o *object) isRootVar() bool {
	if o == nil {
		return false
	}

	if o.rootVar == nil {
		return true
	}

	return false
}

func (o *object) isPtr() bool {
	return o != nil && o.isPointer
}

func (o *object) isStruct() bool {
	if o == nil {
		return false
	}

	return isStruct(o.obj)
}

func isStruct(obj types.Object) bool {
	_, is := obj.(*types.TypeName)

	return is
}

func (v *visiter) ObjUnarExported(ident *ast.Ident) *object {
	return v.objUnar(ident, true)
}

func (v *visiter) ObjUnar(ident *ast.Ident) *object {
	return v.objUnar(ident, false)
}

func (v *visiter) objUnar(ident *ast.Ident, isForExported bool) *object {
	obj := v.pass.TypesInfo.ObjectOf(ident)
	parent := v.getReadOnly(obj)

	var parentIdent *ast.Ident
	if parent != nil {
		parentIdent = parent.ident
	}

	return v.obj(ident, parentIdent, false, isForExported)
}

func (v *visiter) Obj(ident, pIdent *ast.Ident) *object {
	return v.obj(ident, pIdent, false, false)
}

func (v *visiter) ObjPtr(ident, pIdent *ast.Ident) *object {
	return v.obj(ident, pIdent, true, false)
}

func (v *visiter) obj(
	ident, pIdent *ast.Ident,
	isPtr bool,
	isForExported bool,
) *object {
	parentObj := v.pass.TypesInfo.ObjectOf(pIdent)
	parent := v.getReadOnly(parentObj)

	rootObj := parentObj

	var rootVarObj types.Object

	if !isStruct(parentObj) {
		rootVarObj = parentObj
	}

	if parent != nil {
		if parent.root != nil && v.isReadOnly(parent.root) {
			rootObj = v.getReadOnly(parent.root).obj
		}

		if parent.rootVar != nil && v.isReadOnly(parent.rootVar) {
			rootVarObj = v.getReadOnly(parent.rootVar).obj
		}
	}

	obj := v.pass.TypesInfo.ObjectOf(ident)

	return &object{
		ident:         ident,
		obj:           obj,
		parent:        parentObj,
		root:          rootObj,
		rootVar:       rootVarObj,
		isPointer:     isPtr || parent.isPtr(),
		isForExported: isForExported,
	}
}

func (v *visiter) walk(n ast.Node) {
	if n != nil {
		ast.Walk(v, n)
	}
}

// TODO fix dependencies on the order of files within the package.
//
//nolint:nestif,funlen,gocognit,gocyclo,cyclop // FIXME
func (v *visiter) Visit(node ast.Node) ast.Visitor {
	// TODO remove debug information
	line := 0

	if node != nil {
		pos := v.pass.Fset.Position(node.Pos())
		if pos.IsValid() {
			line = pos.Line
		}
	}

	_ = line

	// Mark structs
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

		// struct.SomeField
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

	// global variable declaration in file
	valueSpec, ok := node.(*ast.ValueSpec)
	if ok {
		// TODO multiple assign
		if len(valueSpec.Names) == 0 {
			return v
		}

		lFirst := valueSpec.Names[0]

		lIdent := lFirst
		v.markROByName(lIdent)

		rFirst := valueSpec.Values[0]
		isPtrRight := false

		if unaryExpr, ok := rFirst.(*ast.UnaryExpr); ok &&
			unaryExpr.Op == token.AND {
			rFirst = unaryExpr.X
			isPtrRight = true
		}

		if compositeList, ok := rFirst.(*ast.CompositeLit); ok {
			rFirst = compositeList.Type
		}

		if rIdent, ok := rFirst.(*ast.Ident); ok {
			v.markROByRight(lIdent, rIdent, isPtrRight)
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

	starExpr, isPtrLeft := lhsFirst.(*ast.StarExpr)
	if isPtrLeft {
		lhsFirst = starExpr.X
	}

	// structs
	lSelectorExpr, ok := lhsFirst.(*ast.SelectorExpr)
	if ok {
		if ident, structNamed, ok := v.getStructNamed(lSelectorExpr); ok {
			obj := structNamed.Obj()
			if v.isReadOnly(obj) {
				ro := v.getReadOnly(obj)
				v.tryMutate = append(v.tryMutate, v.Obj(ident, ro.ident))
			}

			lIdent = ident
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
			if !ro.isPtr() ||
				ro.isPtr() && ro.isRootVar() ||
				ro.isPtr() && isPtrLeft {
				v.tryMutate = append(v.tryMutate, v.Obj(lIdent, ro.ident))
			}

			return v
		}
	}

	// TODO replace on more useful configuration
	// Mark variable
	v.markROByName(lIdent)

	// read Rhs
	rFirst := assign.Rhs[0]
	isPtrRight := false

	if unaryExpr, ok := rFirst.(*ast.UnaryExpr); ok &&
		unaryExpr.Op == token.AND {
		rFirst = unaryExpr.X
		isPtrRight = true
	}

	if selectorExpr, ok := rFirst.(*ast.SelectorExpr); ok {
		rFirst = selectorExpr.Sel
	}

	if compositeList, ok := rFirst.(*ast.CompositeLit); ok {
		rFirst = compositeList.Type
	}

	if rIdent, ok := rFirst.(*ast.Ident); ok {
		v.markROByRight(lIdent, rIdent, isPtrRight)
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

	if strings.HasSuffix(ident.Name, "ROExt") {
		obj := v.pass.TypesInfo.ObjectOf(ident)
		v.readonlyObjects[obj] = v.ObjUnarExported(ident)
	}
}

func (v *visiter) markROByRight(lIdent, rIdent *ast.Ident, isPtr bool) {
	lObj := v.pass.TypesInfo.ObjectOf(lIdent)
	rObj := v.pass.TypesInfo.ObjectOf(rIdent)

	ro := v.getReadOnly(rObj)
	if v.isReadOnly(rObj) &&
		(ro.isPtr() || ro.isStruct() || isPtr) {
		v.readonlyObjects[lObj] = v.obj(lIdent, rIdent, isPtr, false)
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
