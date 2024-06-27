package init

import "context"

type GenericStructWithInit[T any] struct { // want GenericStructWithInit:"<init:own>"
	Field T
}

func (g GenericStructWithInit[T]) Init(ctx context.Context) error { // want "Init method is not declared on pointer receiver"
	return nil
}

type EmbeddedGeneric struct { // want EmbeddedGeneric:"<init:GenericStructWithInit>"
	GenericStructWithInit[string] /* want
	"type testdata/src/init.GenericStructWithInit\\[string\\] has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	"Init method of GenericStructWithInit will be called 2 times"
	*/
}

type GenericInit[T any] struct{} // want GenericInit:"<init:none>"

func (g GenericInit[T]) Init(t T) error {
	return nil
}

type EmbeddedGenericInitCtx struct { // want EmbeddedGenericInitCtx:"<init:GenericInit>"
	GenericInit[context.Context] /* want
	"type testdata/src/init.GenericInit\\[context.Context\\] has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	"Init method of GenericInit will be called 2 times"
	*/
}
