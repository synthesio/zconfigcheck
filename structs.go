package zconfigcheck

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"math"
	"strings"

	"github.com/fatih/structtag"
)

const (
	keyTag         = "key"
	defaultTag     = "default"
	descriptionTag = "description"
	injectTag      = "inject"
	injectAsTag    = "inject-as"
)

type structFact struct {
	Issues   []string
	InitPath string
	InitPos  token.Pos
}

func (structFact) AFact() {}

func (s structFact) String() string {
	if s.InitPos == token.NoPos {
		return "<init:none>"
	}

	if s.InitPath == "" {
		return "<init:own>"
	}

	return fmt.Sprintf("<init:%s>", s.InitPath)
}

// TypeSet is an implementation of an ordered set. Its main purpose
// is detecting dependency cycles between types.
type TypeSet []types.Type

func (ts TypeSet) String() string {
	nodes := make([]string, 0, len(ts))

	for _, typ := range ts {
		nodes = append(nodes, typ.String())
	}

	return strings.Join(nodes, " -> ")
}

// Add returns a copy of the current set plus the given Type argument.
// The original set (the receiver) is not modified.
// If the given argument already belongs to the set, an error is returned.
func (ts TypeSet) Add(t types.Type) (TypeSet, error) {
	newSet := make(TypeSet, 0, len(ts)+1)

	var hasCycle bool
	for _, typ := range ts {
		newSet = append(newSet, typ)

		if typ == t {
			hasCycle = true
		}
	}

	newSet = append(newSet, t)
	if hasCycle {
		return nil, errors.New(newSet.String())
	}

	return newSet, nil
}

// checkStructs visits the AST to find all struct declaration.
// All found structs are analyzed and any issues are reported.
func (c *checker) checkStructs() {
	c.Inspector.Preorder([]ast.Node{(*ast.TypeSpec)(nil), (*ast.StructType)(nil)}, func(node ast.Node) {
		switch n := node.(type) {
		// these are named type declarations
		case *ast.TypeSpec:
			obj := c.Pass.TypesInfo.ObjectOf(n.Name)

			issues := c.checkStruct(obj.Type(), obj)

			// do not report issues on struct fields if this is an alias (e.g. type MyType MyOtherType)
			// this ensures that the same issues are not reported more than once
			rhsType := c.Pass.TypesInfo.TypeOf(n.Type)
			if _, ok := rhsType.(*types.Struct); ok {
				issues.Report(c.Pass)
			}
		// this is necessary to catch anonymous struct declarations
		case *ast.StructType:
			typ := c.Pass.TypesInfo.TypeOf(n)

			// This could be a StructType node descending from a TypeSpec node.
			// If that's the case, since we're doing a Preorder visit we will already
			// have visited the TypeSpec parent and the type will already be known.
			if _, ok := c.PkgStructs[typ]; ok {
				return
			}

			issues := c.checkStruct(typ, nil)
			issues.Report(c.Pass)
		}
	})
}

// checkStruct returns any detected issues with the given Type.
// If the argument is not a struct, then no issue is returned.
// All struct types are also registered in the PkgStruct map.
func (c *checker) checkStruct(typ types.Type, obj types.Object) Issues {
	str, ok := typ.Underlying().(*types.Struct)
	if !ok {
		return nil
	}

	info := c.parseStruct(str, typ, nil)

	// This is done outside the recursive parseStruct method because
	// we only want to report issues with the Init method of the root struct.
	// Any issues with child structs are reported in detail when those structs are eventually visited.
	info.Issues = info.Issues.Merge(info.InitIssues)

	c.PkgStructs[typ] = info
	c.PkgStructs[typ.Underlying()] = info
	if obj != nil {
		c.Pass.ExportObjectFact(obj, info.Fact())
	}

	// we have a fully built scope for this struct, so we can check any issues
	return info.Issues.Merge(info.Scope.Check())
}

// parseStruct recursively visits the directed graph defined by the struct and its fields.
// It returns a StructInfo type containing all collected information about the visited struct.
// The recursive visit stops early when a cyclic dependency is detected.
func (c *checker) parseStruct(str *types.Struct, typ types.Type, set TypeSet) StructInfo {
	info := StructInfo{
		Scope:      NewScope(),
		Issues:     make(Issues),
		InitIssues: make(Issues),
	}

	set, err := set.Add(typ)
	if err != nil {
		info.DependencyCycles = append(info.DependencyCycles, err.Error())
		return info
	}

	for i := 0; i < str.NumFields(); i++ {
		strField := str.Field(i)
		rawTags := str.Tag(i)

		if !strField.Exported() {
			if rawTags != "" {
				info.Issues.Add(strField.Pos(), "private fields cannot have tags")
			}
			continue
		}

		// collect all necessary information about the current struct field
		field := newStructField(strField, i)

		tags, issues := parseTags(rawTags)
		info.Issues.Add(strField.Pos(), issues...)

		if key, ok := tags[keyTag]; ok {
			field.Key = key
			if _, ok := tags[injectTag]; ok {
				info.Issues.Add(strField.Pos(), "key and inject tags should not be used on the same field")
			}
		} else {
			if _, ok := tags[defaultTag]; ok {
				info.Issues.Add(strField.Pos(), "default tag is used on field without key tag")
			}

			if _, ok := tags[descriptionTag]; ok {
				info.Issues.Add(strField.Pos(), "description tag is used on field without key tag")
			}
		}

		if injectAs, ok := tags[injectAsTag]; ok {
			if !field.IsPointer && !field.IsGeneric {
				info.Issues.Add(strField.Pos(), "field type is not a pointer, cannot be used as injection source")
			}

			field.Alias = injectAs
			field.IsSource = true
		}

		if inject, ok := tags[injectTag]; ok {
			if !field.IsPointer && !field.IsInterface {
				info.Issues.Add(strField.Pos(), "field type is not a pointer nor interface, cannot be used as injection target")
			}

			if field.IsSource {
				field.IsSource = false
				info.Issues.Add(strField.Pos(), "inject and inject-as tags cannot be used on the same field")
			}

			field.Alias = inject
			field.IsTarget = true
			info.Scope.AddTarget(field)
		}

		if field.IsSource {
			info.Scope.AddSource(field)
		}

		if !field.IsStruct() {
			if field.Key != "" {
				info.Scope.AddKey(field)
			}

			// this field is not a struct, so there is no need to visit it
			continue
		}

		if field.IsTarget {
			// if the struct is an injection target, then its declared keys and dependencies
			// will have no effect here
			continue
		}

		fieldInfo := c.parseStruct(field.Struct, field.StructType, set)
		child := ChildInfo{
			StructField: field,
			StructInfo:  fieldInfo,
		}
		info.Children = append(info.Children, child)
		info.DependencyCycles = append(info.DependencyCycles, fieldInfo.DependencyCycles...)

		if field.Key != "" && len(fieldInfo.Scope.Keys) == 0 {
			// this struct has an associated key tag, and it has no tagged fields
			// zconfig will consider it as a leaf, so we can add its key
			info.Scope.AddKey(field)
		}

		for _, issues := range fieldInfo.Issues {
			info.Issues.Add(strField.Pos(), issues...)
		}

		info.MergeScopes(child)
	}

	info.resolveInit(typ, c.Pass.Pkg)

	return info
}

func parseTags(rawTags string) (map[string]string, []string) {
	parsed, err := structtag.Parse(rawTags)
	if err != nil {
		return nil, []string{err.Error()}
	}

	var issues []string
	tags := make(map[string]string)

	for _, key := range []string{keyTag, descriptionTag, injectTag, injectAsTag} {
		tag, err := parsed.Get(key)
		if err != nil {
			continue
		}

		if tag.Name == "" {
			issues = append(issues, fmt.Sprintf("%s tag cannot be empty", key))
			continue
		}

		tags[key] = tag.Name
	}

	tag, err := parsed.Get(defaultTag)
	if err == nil {
		tags[defaultTag] = tag.Name
	}

	return tags, issues
}

type ChildInfo struct {
	StructField
	StructInfo
}

type StructInfo struct {
	Children         []ChildInfo
	Issues           Issues
	Scope            Scope
	DependencyCycles []string

	InitPos       token.Pos
	HasInitMethod bool
	CallCount     int
	InitPath      string
	InitDepth     int
	InitIssues    Issues
}

func (s StructInfo) HasInit() bool {
	return s.InitPos != token.NoPos
}

// Fact converts StructInfo into a structFact. StructInfo does not implement
// the Fact interface because it has references to types which cannot be encoded by
// the gob package, leading to issues with golangci-lint cache.
func (s StructInfo) Fact() *structFact {
	var issues []string
	for alias, paths := range s.Scope.UnresolvedTargets() {
		issues = append(issues, fmt.Sprintf(
			"no source is provided for alias '%s' used by target fields: %s",
			alias, strings.Join(paths, ", ")))
	}

	for _, cycle := range s.DependencyCycles {
		issues = append(issues, fmt.Sprintf("configured struct contains dependency cycle: %s", cycle))
	}

	for _, fieldIssues := range s.Issues {
		issues = append(issues, fieldIssues...)
	}

	return &structFact{
		Issues:   issues,
		InitPath: s.InitPath,
		InitPos:  s.InitPos,
	}
}

// MergeScopes merges the given child's scope into the receiver's one.
// Any issues detected during the merge operation are stored into the receiver issues collection.
func (s *StructInfo) MergeScopes(child ChildInfo) {
	field := child.StructField
	o := child.Scope

	mergeKeys := field.Key != "" || field.IsEmbedded
	if mergeKeys {
		for _, keyFields := range o.Keys {
			for _, keyField := range keyFields {
				keyField.Path = field.Path + "." + keyField.Path
				if field.Key != "" {
					keyField.Key = field.Key + "." + keyField.Key
				}

				keyField.Pos = field.Pos
				s.Scope.AddKey(keyField)
			}
		}
	}

	if !mergeKeys && len(o.Keys) > 0 {
		s.Issues.Add(field.Pos, fmt.Sprintf("field %s contains key tags but is not tagged with a key", field.Path))
	}

	for _, sourceFields := range o.Sources {
		for _, sourceField := range sourceFields {
			sourceField.Path = field.Path + "." + sourceField.Path
			sourceField.Pos = field.Pos

			s.Scope.AddSource(sourceField)
		}
	}

	for _, targets := range o.Targets {
		for _, target := range targets {
			target.Path = field.Path + "." + target.Path
			target.Pos = field.Pos

			s.Scope.AddTarget(target)
		}
	}
}

// lookupInitMethod checks whether the given type implements the zconfig.Initializable interface.
// It returns two parameters:
// - the position of the method (it will be token.NoPos if no method is found)
// - a boolean indicating whether the method is implemented on a pointer receiver
func lookupInitMethod(typ types.Type, pkg *types.Package) (token.Pos, bool) {
	init, _, _ := types.LookupFieldOrMethod(types.NewPointer(typ), true, pkg, "Init")
	if init == nil {
		return 0, false
	}

	fn, ok := init.Type().(*types.Signature)
	if !ok {
		return 0, false
	}

	switch fn.Params().Len() {
	case 0:
	case 1:
		if fn.Params().At(0).Type().String() != "context.Context" {
			return 0, false
		}
	default:
		return 0, false
	}

	if fn.Results().Len() != 1 ||
		fn.Results().At(0).Type().String() != "error" ||
		!sameNamedType(fn.Recv().Type(), typ) {
		return 0, false
	}
	_, isPtr := fn.Recv().Type().(*types.Pointer)
	return init.Pos(), isPtr
}

func sameNamedType(t1, t2 types.Type) bool {
	if ptr, ok := t1.(*types.Pointer); ok {
		t1 = ptr.Elem()
	}
	if ptr, ok := t2.(*types.Pointer); ok {
		t2 = ptr.Elem()
	}

	named1, ok := t1.(*types.Named)
	if !ok {
		return false
	}

	named2, ok := t2.(*types.Named)
	if !ok {
		return false
	}

	return named1.Obj() == named2.Obj()
}

// resolveInit scans the struct for possible issues with its own zconfig.Initializable implementation and
// that of its fields.
// Any detected issues are stored in the InitIssues collection, to avoid reporting them more than once when
// a struct is used as field of another struct.
func (s *StructInfo) resolveInit(typ types.Type, pkg *types.Package) {
	pos, isPtr := lookupInitMethod(typ, pkg)
	if pos != token.NoPos {
		// an Init method is implemented using this struct (or a pointer to it) as its receiver
		s.HasInitMethod = true
		s.InitPos = pos

		if !isPtr {
			// if the Init method is not implemented on a pointer receiver, then it will always be called
			s.CallCount = 1
			s.InitIssues.Add(pos, "Init method is not declared on pointer receiver")
		}
	}

	var embedded []ChildInfo

	// We need to track the embedding depth, meaning how many embeds are done from the current struct
	// to the one implementing an Init method.
	// If no embed is present or none of them implement an Init method, then the index will be -1
	minDepth := math.MaxInt
	minDepthIndex := -1

	for i, child := range s.Children {
		if !child.HasInit() {
			continue
		}

		if len(child.InitIssues) > 0 {
			// Report a generic message on the field if it has Init issues. Detailed issues will already be
			// reported on that struct fields.
			s.InitIssues.Add(child.Pos, fmt.Sprintf(
				"type %s has one or more issues with Init methods implemented by itself, embedded structs or its fields",
				child.StructType.String(),
			))
		}

		// If the field is a pointer, then we know that its Init method will be called
		if child.CallCount == 0 && child.IsPointer {
			child.CallCount = 1
			s.Children[i] = child
		}

		path := child.Path
		if child.InitPath != "" {
			path += "." + child.InitPath
		}

		if !child.IsEmbedded {
			// If this is not an embedded field, then zconfig cannot have access to a pointer so its Init method
			// won't be called
			if child.CallCount == 0 {
				s.InitIssues.Add(child.Pos, fmt.Sprintf("Init method of %s won't be called", path))
			}

			continue
		}

		child.InitPath = path
		embedded = append(embedded, child)

		if child.InitDepth > minDepth {
			continue
		}

		// this means that we already found an embed at the same depth before, thus
		// there is an ambiguity and this struct does not inherit any of the Init methods
		if child.InitDepth == minDepth {
			minDepthIndex = -1
			continue
		}

		minDepthIndex = len(embedded) - 1
		minDepth = child.InitDepth
	}

	if len(embedded) == 0 {
		return
	}

	// if the struct has its own Init method, then no Init method is inherited
	if s.HasInitMethod {
		minDepthIndex = -1
	}

	if minDepthIndex != -1 {
		e := embedded[minDepthIndex]
		s.InitPos = e.InitPos
		s.InitDepth = e.InitDepth + 1
		s.InitPath = e.InitPath

		if e.CallCount > 0 {
			// The Init method will already be called once on the embedded field. Since it is callable,
			// it will also be called on the embedding struct, so its call count increases by 1.
			s.CallCount = e.CallCount + 1
		}
		if s.CallCount > 1 {
			s.InitIssues.Add(e.Pos, fmt.Sprintf("Init method of %s will be called %d times", s.InitPath, s.CallCount))
		}
	}

	for i, e := range embedded {
		// The Init method implemented by the embedded struct is not inherited by the embedding one.
		// Since we know it cannot be called on the embedded field, then we know it will never be called
		if i != minDepthIndex && e.CallCount == 0 {
			s.InitIssues.Add(e.Pos, fmt.Sprintf("Init method of %s won't be called", e.InitPath))
		}
	}
}

type StructField struct {
	typeOrConstraint types.Type

	Struct     *types.Struct
	StructType types.Type
	Index      int

	Pos         token.Pos
	IsGeneric   bool
	IsEmbedded  bool
	IsPointer   bool
	IsInterface bool

	Key      string
	Alias    string
	IsSource bool
	IsTarget bool
	Path     string
}

func (s StructField) String() string {
	return fmt.Sprintf("%s %s", s.Path, s.typeOrConstraint.String())
}

// CompatibleWith returns true if the two struct fields have compatible types or constraints.
// This means that a unique source can be assigned to both target fields.
// Example:
// - field a is of type int
// - field b is of type T[constraints.Integer]
// They are compatible, because they can both be assigned an int type
func (s StructField) CompatibleWith(f StructField) bool {
	return types.AssignableTo(s.typeOrConstraint, f.typeOrConstraint) ||
		types.AssignableTo(f.typeOrConstraint, s.typeOrConstraint)
}

// AssignableTo returns true if the method receiver can be assigned to the dest argument.
// If the method receiver is a generic type, then we consider it assignable to avoid
// raising potentially false issues.
func (s StructField) AssignableTo(dest StructField) bool {
	if s.IsGeneric && !dest.IsGeneric {
		return true
	}
	return types.AssignableTo(s.typeOrConstraint, dest.typeOrConstraint)
}

func (s StructField) IsStruct() bool {
	return s.Struct != nil
}

func newStructField(field *types.Var, index int) StructField {
	typ := field.Type()
	ptr, isPtr := typ.(*types.Pointer)
	if isPtr {
		typ = ptr.Elem()
	}

	typeParam, _ := typ.(*types.TypeParam)
	typeOrConstraint := field.Type()
	if typeParam != nil {
		typeOrConstraint = typeParam.Constraint()
	}

	s := StructField{
		typeOrConstraint: typeOrConstraint,
		Index:            index,
		Pos:              field.Pos(),
		IsGeneric:        typeParam != nil,
		IsPointer:        isPtr,
		IsEmbedded:       field.Embedded(),
		Path:             field.Name(),
	}

	switch x := typ.Underlying().(type) {
	case *types.Interface:
		s.IsInterface = true
	case *types.Struct:
		s.Struct = x
		s.StructType = typ
	}

	return s
}
