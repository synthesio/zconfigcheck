package init

import (
	"context"
	"fmt"
)

type DetectCalls struct { // want DetectCalls:"<init:own>"
	*EmbeddedPtr // want "type testdata/src/init.EmbeddedPtr has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	WithCtx      // want "Init method of WithCtx won't be called"
	Field        *NoCtx
	Field2       NoCtx                        // want "Init method of Field2 won't be called"
	Field3       NonPtrReceiver               // want "type testdata/src/init.NonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	Field4       *NonPtrReceiver              // want "type testdata/src/init.NonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	Generic1     GenericStructWithInit[bool]  // want "type testdata/src/init.GenericStructWithInit\\[bool\\] has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	Generic2     GenericInit[context.Context] // want "type testdata/src/init.GenericInit\\[context.Context\\] has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	Generic3     GenericInit[bool]
}

func (d *DetectCalls) Init() error {
	d.EmbeddedPtr.Init(context.Background())              // want "Init method is already invoked by zconfig"
	fmt.Println(d.EmbeddedPtr.Init(context.Background())) // want "Init method is already invoked by zconfig"

	var f WithCtx
	f.Init(context.Background())

	Init(context.Background())

	go d.Field.Init() // want "Init method is already invoked by zconfig"

	defer func() {
		d.Field.Init() // want "Init method is already invoked by zconfig"

		d.initHelper()
	}()

	d.Generic1.Init(context.Background()) // want "Init method is already invoked by zconfig"
	d.Generic2.Init(context.Background()) // want "Init method is already invoked by zconfig"

	callInit(d)

	d.Init() // want "Init method is already invoked by zconfig"

	a := d
	a.Init()       // want "Init method is already invoked by zconfig"
	a.Field.Init() // want "Init method is already invoked by zconfig"

	d.Field2.Init()
	d.Field3.Init(context.Background()) // want "Init method is already invoked by zconfig"
	d.Field4.Init(context.Background()) // want "Init method is already invoked by zconfig"

	return d.WithCtx.Init(context.Background())
}

func (d *DetectCalls) initHelper() {
	d.Field.Init() // want "Init method is already invoked by zconfig"
}

func callInit(d *DetectCalls) {
	d.Field.Init() // want "Init method is already invoked by zconfig"
	d.Init()       // want "Init method is already invoked by zconfig"
}

func Init(ctx context.Context) error {
	return nil
}
