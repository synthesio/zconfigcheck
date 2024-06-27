package init

import (
	"testdata/src/init/subpackage"
)

type EmbeddedNonPtrReceiver struct { // want EmbeddedNonPtrReceiver:"<init:NonPtrReceiver>"
	NonPtrReceiver /* want
	"Init method of NonPtrReceiver will be called 2 times"
	"type testdata/src/init.NonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	*/
}

type EmbeddedNonPtrReceiver2 struct { // want EmbeddedNonPtrReceiver2:"<init:NonPtrReceiver>"
	*NonPtrReceiver /* want
	"Init method of NonPtrReceiver will be called 2 times"
	"type testdata/src/init.NonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	*/
}

type EmbeddedWithInit struct { // want EmbeddedWithInit:"<init:own>"
	NonPtrReceiver /* want
	"type testdata/src/init.NonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	*/
}

func (EmbeddedWithInit) Init() error { return nil } // want "Init method is not declared on pointer receiver"

type EmbeddedWithInit2 struct { // want EmbeddedWithInit2:"<init:own>"
	*NonPtrReceiver /* want
	"type testdata/src/init.NonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	*/
}

func (*EmbeddedWithInit2) Init() error { return nil }

type EmbeddedInit struct { // want EmbeddedInit:"<init:WithCtx>"
	WithCtx
}

type EmbeddedPtr struct { // want EmbeddedPtr:"<init:WithCtx>"
	*WithCtx // want "Init method of WithCtx will be called 2 times"
}

type EmbeddedOtherPkg struct { // want EmbeddedOtherPkg:"<init:Init>"
	subpackage.Init
}

type EmbeddedOtherPkg2 struct { // want EmbeddedOtherPkg2:"<init:Init>"
	*subpackage.Init // want "Init method of Init will be called 2 times"
}

type EmbedMany struct { // want EmbedMany:"<init:none>"
	WithCtx // want "Init method of WithCtx won't be called"
	NoCtx   // want "Init method of NoCtx won't be called"
}

type EmbedMany2 struct { // want EmbedMany2:"<init:none>"
	*WithCtx
	NoCtx // want "Init method of NoCtx won't be called"
}

type EmbedMany3 struct { // want EmbedMany3:"<init:none>"
	WithCtx // want "Init method of WithCtx won't be called"
	*NoCtx
}

type EmbedMany4 struct { // want EmbedMany4:"<init:none>"
	*WithCtx
	*NoCtx
}

type EmbedManyWithInit struct { // want EmbedManyWithInit:"<init:own>"
	WithCtx // want "Init method of WithCtx won't be called"
	NoCtx   // want "Init method of NoCtx won't be called"
}

func (*EmbedManyWithInit) Init() error { return nil }

type EmbedManyWithInit2 struct { // want EmbedManyWithInit2:"<init:own>"
	*WithCtx
	NoCtx // want "Init method of NoCtx won't be called"
}

func (*EmbedManyWithInit2) Init() error { return nil }

type EmbedManyWithInit3 struct { // want EmbedManyWithInit3:"<init:own>"
	WithCtx // want "Init method of WithCtx won't be called"
	*NoCtx
}

func (*EmbedManyWithInit3) Init() error { return nil }

type EmbedManyWithInit4 struct { // want EmbedManyWithInit4:"<init:own>"
	*WithCtx
	*NoCtx
}

func (*EmbedManyWithInit4) Init() error { return nil }

type EmbedMany9 struct { // want EmbedMany9:"<init:WithCtx>"
	WithCtx
	*EmbeddedPtr /* want
	"type testdata/src/init.EmbeddedPtr has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	*/
}

type EmbedMany10 struct { // want EmbedMany10:"<init:WithCtx>"
	WithCtx
	*EmbeddedNonPtrReceiver /* want
	"type testdata/src/init.EmbeddedNonPtrReceiver has one or more issues with Init methods implemented by itself, embedded structs or its fields"
	*/
}

type EmbedMany11 struct { // want EmbedMany11:"<init:EmbedMany9.WithCtx>"
	EmbedMany9 // want "type testdata/src/init.EmbedMany9 has one or more issues with Init methods implemented by itself, embedded structs or its fields"
}
