package init

import (
	"context"
)

type WithCtx struct{}                           // want WithCtx:"<init:own>"
func (*WithCtx) Init(ctx context.Context) error { return nil }

type NoCtx struct{}        // want NoCtx:"<init:own>"
func (*NoCtx) Init() error { return nil }

type NonPtrReceiver struct{} // want NonPtrReceiver:"<init:own>"

func (NonPtrReceiver) Init(ctx context.Context) error { // want "Init method is not declared on pointer receiver"
	return nil
}

type EmbedNonPtrReceiver struct { // want EmbedNonPtrReceiver:"<init:NonPtrReceiver>"
	NonPtrReceiver /* want
	"type testdata/src/init.NonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	"Init method of NonPtrReceiver will be called 2 times"
	*/
}

type PtrEmbedNonPtrReceiver struct { // want PtrEmbedNonPtrReceiver:"<init:NonPtrReceiver>"
	*NonPtrReceiver /* want
	"type testdata/src/init.NonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	"Init method of NonPtrReceiver will be called 2 times"
	*/
}

type PtrReceiver struct{} // want PtrReceiver:"<init:own>"

func (*PtrReceiver) Init(ctx context.Context) error {
	return nil
}

type EmbedPtrReceiver struct { // want EmbedPtrReceiver:"<init:PtrReceiver>"
	PtrReceiver
}

type PtrEmbedPtrReceiver struct { // want PtrEmbedPtrReceiver:"<init:PtrReceiver>"
	*PtrReceiver // want "Init method of PtrReceiver will be called 2 times"
}

type InitializableFields struct { // want InitializableFields:"<init:none>"
	Called  NonPtrReceiver  // want "type testdata/src/init.NonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	Called2 *NonPtrReceiver // want "type testdata/src/init.NonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"

	Called3   *PtrReceiver
	NotCalled PtrReceiver // want "Init method of NotCalled won't be called"

	Called4 PtrEmbedPtrReceiver  // want "type testdata/src/init.PtrEmbedPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	Called5 *PtrEmbedPtrReceiver // want "type testdata/src/init.PtrEmbedPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"

	NotCalled2 EmbedPtrReceiver // want "Init method of NotCalled2.PtrReceiver won't be called"
	Called6    *EmbedPtrReceiver

	Called7 EmbedNonPtrReceiver /* want
	"type testdata/src/init.EmbedNonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	*/
	Called8 *EmbedNonPtrReceiver /* want
	"type testdata/src/init.EmbedNonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	*/
	Called9 PtrEmbedNonPtrReceiver /* want
	"type testdata/src/init.PtrEmbedNonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	*/
	Called10 *PtrEmbedNonPtrReceiver /* want
	"type testdata/src/init.PtrEmbedNonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	*/
}

type Embed2 struct { // want Embed2:"<init:EmbedNonPtrReceiver.NonPtrReceiver>"
	EmbedNonPtrReceiver /* want
	"type testdata/src/init.EmbedNonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	"Init method of EmbedNonPtrReceiver.NonPtrReceiver will be called 3 times"
	*/
}

type Embed3 struct { // want Embed3:"<init:EmbedNonPtrReceiver.NonPtrReceiver>"
	*EmbedNonPtrReceiver /* want
	"type testdata/src/init.EmbedNonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	"Init method of EmbedNonPtrReceiver.NonPtrReceiver will be called 3 times"
	*/
}

type Embed4 struct { // want Embed4:"<init:PtrEmbedNonPtrReceiver.NonPtrReceiver>"
	PtrEmbedNonPtrReceiver /* want
	"type testdata/src/init.PtrEmbedNonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	"Init method of PtrEmbedNonPtrReceiver.NonPtrReceiver will be called 3 times"
	*/
}

type Embed5 struct { // want Embed5:"<init:PtrEmbedNonPtrReceiver.NonPtrReceiver>"
	*PtrEmbedNonPtrReceiver /* want
	"type testdata/src/init.PtrEmbedNonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	"Init method of PtrEmbedNonPtrReceiver.NonPtrReceiver will be called 3 times"
	*/
}

type Embed6 struct { // want Embed6:"<init:EmbedPtrReceiver.PtrReceiver>"
	EmbedPtrReceiver
}

type Embed7 struct { // want Embed7:"<init:EmbedPtrReceiver.PtrReceiver>"
	*EmbedPtrReceiver /* want
	"Init method of EmbedPtrReceiver.PtrReceiver will be called 2 times"
	*/
}

type Embed8 struct { // want Embed8:"<init:PtrEmbedPtrReceiver.PtrReceiver>"
	PtrEmbedPtrReceiver /* want
	"type testdata/src/init.PtrEmbedPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	"Init method of PtrEmbedPtrReceiver.PtrReceiver will be called 3 times"
	*/
}

type Embed9 struct { // want Embed9:"<init:PtrEmbedPtrReceiver.PtrReceiver>"
	*PtrEmbedPtrReceiver /* want
	"type testdata/src/init.PtrEmbedPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	"Init method of PtrEmbedPtrReceiver.PtrReceiver will be called 3 times"
	*/
}
