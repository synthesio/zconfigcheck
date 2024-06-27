package injection

import (
	"context"

	"github.com/synthesio/zconfig/v2"
	"testdata/src/injection/subpackage"
)

type root struct { // want root:"<init:none>"
	// assert that anonymous structs are correctly processed
	Node1 struct { /* want
		"injection alias 'leaf': cannot inject source field 'Node1.Leaf \\*bool' into target field 'Leaf \\*string', mismatched types"
		*/
		Leaf *bool `inject-as:"leaf"`
	}

	Node2 node // want "injection alias 'leaf': target fields 'Node2.Leaf \\*bool' and 'Leaf \\*string' are incompatible, mismatched types"

	Leaf *string `inject:"leaf"` /* want
	"injection alias 'leaf': target fields 'Leaf \\*string' and 'Node2.Leaf \\*bool' are incompatible, mismatched types"
	"injection alias 'leaf': target field 'Leaf \\*string' cannot be injected with source field 'Node1.Leaf \\*bool', mismatched types"
	*/
}

type node struct { // want node:"<init:none>"
	Leaf *bool `inject:"leaf"`
}

// These structs assert that scopes belonging to structs imported from other packages are correctly processed
type duplicateSources struct { // want duplicateSources:"<init:none>"
	Src1 subpackage.InjectionSrc // want "inject-as alias 'source' defined by field Src1.Source is already used by field Src2.Source"
	Src2 subpackage.InjectionSrc // want "inject-as alias 'source' defined by field Src2.Source is already used by field Src1.Source"
}

type incompatibleTypes struct { // want incompatibleTypes:"<init:none>"
	Source subpackage.InjectionSrc /* want
	"injection alias 'source': cannot inject source field 'Source.Source \\*string' into target field 'Target.Target \\*int', mismatched types"
	*/
	Target subpackage.WrongInjectionTarget /* want
	"injection alias 'source': target fields 'Target.Target \\*int' and 'Target2.Target \\*string' are incompatible, mismatched types"
	"injection alias 'source': target field 'Target.Target \\*int' cannot be injected with source field 'Source.Source \\*string', mismatched types"
	*/
	Target2 subpackage.InjectionTarget /* want
	"injection alias 'source': target fields 'Target2.Target \\*string' and 'Target.Target \\*int' are incompatible, mismatched types"
	*/
}

// No error expected, injecting an implementation into an interface is allowed
type interfaces struct { // want interfaces:"<init:none>"
	Source *subpackage.DoerImpl `inject-as:"itf"`
	Target subpackage.Doer      `inject:"itf"`
}

type missingSources struct { // want missingSources:"<init:none>"
	Target  subpackage.InjectionTarget
	Target2 subpackage.InjectionTarget
	Target3 *bool `inject:"target3"`
}

var anon struct {
	Target subpackage.InjectionTarget
}

func main() {
	Configure(context.Background(), new(missingSources)) /* want
	"no source is provided for alias 'target3' used by target fields: Target3"
	"no source is provided for alias 'source' used by target fields: Target.Target, Target2.Target"
	*/

	var missing missingSources
	Configure(context.Background(), &missing) /* want
	"no source is provided for alias 'target3' used by target fields: Target3"
	"no source is provided for alias 'source' used by target fields: Target.Target, Target2.Target"
	*/

	var missing2 missingSources
	zconfig.Configure(context.Background(), &missing2) /* want
	"no source is provided for alias 'target3' used by target fields: Target3"
	"no source is provided for alias 'source' used by target fields: Target.Target, Target2.Target"
	*/
}

func Configure(ctx context.Context, str any) { // want Configure:""
	err := zconfig.Configure(ctx, str)
	if err != nil {
		panic(err)
	}
}
