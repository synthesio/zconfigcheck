package injection // want package:"has wrappers"

import (
	"context"

	"github.com/synthesio/zconfig/v2"
)

type GenericSource[T any] struct { // want GenericSource:"<init:none>"
	Source T `inject-as:"t"`
}

type GenericTarget[T any] struct { // want GenericTarget:"<init:none>"
	Target T `inject:"t"`
}

type Generic2[T any, R string] struct { // want Generic2:"<init:none>"
	GenericSource[T] /* want
	"injection alias 't': cannot inject source field 'GenericSource.Source any' into target field 'GenericTarget.Target string', mismatched types"
	*/
	GenericTarget[R] /* want
	"injection alias 't': target field 'GenericTarget.Target string' cannot be injected with source field 'GenericSource.Source any', mismatched types"
	*/
}

type MyG struct { // want MyG:"<init:none>"
	G Generic2[*bool, string] /* want
	"injection alias 't': cannot inject source field 'G.GenericSource.Source \\*bool' into target field 'G.GenericTarget.Target string', mismatched types"
	"injection alias 't': target field 'G.GenericTarget.Target string' cannot be injected with source field 'G.GenericSource.Source \\*bool', mismatched types"
	"field type is not a pointer nor interface, cannot be used as injection target"
	*/
}

type Ok struct { // want Ok:"<init:none>"
	GenericSource[*bool]
	GenericTarget[*bool]
}

type WrongTypeArgs struct { // want WrongTypeArgs:"<init:none>"
	GenericSource[bool] /* want
	"injection alias 't': cannot inject source field 'GenericSource.Source bool' into target field 'GenericTarget.Target string', mismatched types"
	"field type is not a pointer, cannot be used as injection source"
	*/
	GenericTarget[string] /* want
	"injection alias 't': target field 'GenericTarget.Target string' cannot be injected with source field 'GenericSource.Source bool', mismatched types"
	"field type is not a pointer nor interface, cannot be used as injection target"
	*/
}

type WrongTypeArgs2 struct { // want WrongTypeArgs2:"<init:none>"
	GenericTarget[*GenericSource[*bool]] /* want
	"injection alias 't': target field 'GenericTarget.Target \\*testdata/src/injection.GenericSource\\[\\*bool\\]' cannot be injected with source field 'GenericSource.Source \\*testdata/src/injection.GenericTarget\\[\\*bool\\]', mismatched types"
	"injection alias 't': target fields 'GenericTarget.Target \\*testdata/src/injection.GenericSource\\[\\*bool]' and 'GenericSource.Source.Target \\*bool' are incompatible, mismatched types"
	*/
	GenericSource[*GenericTarget[*bool]] /* want
	"injection alias 't': cannot inject source field 'GenericSource.Source \\*testdata/src/injection.GenericTarget\\[\\*bool\\]' into target field 'GenericTarget.Target \\*testdata/src/injection.GenericSource\\[\\*bool\\]', mismatched types"
	"injection alias 't': cannot inject source field 'GenericSource.Source \\*testdata/src/injection.GenericTarget\\[\\*bool\\]' into target field 'GenericSource.Source.Target \\*bool', mismatched types"
	"injection alias 't': target field 'GenericSource.Source.Target \\*bool' cannot be injected with source field 'GenericSource.Source \\*testdata/src/injection.GenericTarget\\[\\*bool\\]', mismatched types"
	"injection alias 't': target fields 'GenericSource.Source.Target \\*bool' and 'GenericTarget.Target \\*testdata/src/injection.GenericSource\\[\\*bool]' are incompatible, mismatched types"
	*/
}

type IncompatibleTargets struct { // want IncompatibleTargets:"<init:none>"
	A GenericTarget[*bool]   // want "injection alias 't': target fields 'A.Target \\*bool' and 'B.Target \\*string' are incompatible, mismatched types"
	B GenericTarget[*string] // want "injection alias 't': target fields 'B.Target \\*string' and 'A.Target \\*bool' are incompatible, mismatched types"
}

type MissingSource struct { // want MissingSource:"<init:none>"
	A GenericTarget[*bool]
}

var _ = zconfig.Configure(context.Background(), new(MissingSource)) // want "no source is provided for alias 't' used by target fields: A.Target"

type InjectFromGeneric struct { // want InjectFromGeneric:"<init:none>"
	A GenericSource[struct { /* want
		"injection alias 'source': cannot inject source field 'A.Source.Source \\*bool' into target field 'B.Source.Target \\*string"
		"inject-as alias 't' defined by field A.Source is already used by field B.Source"
		"field type is not a pointer, cannot be used as injection source"
		*/
		Source *bool `inject-as:"source"`
	}]
	B GenericSource[*struct { /* want
		"inject-as alias 't' defined by field B.Source is already used by field A.Source"
		"injection alias 'source': target field 'B.Source.Target \\*string' cannot be injected with source field 'A.Source.Source \\*bool', mismatched types"
		*/
		Target *string `inject:"source"`
	}]
}
