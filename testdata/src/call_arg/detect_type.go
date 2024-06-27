package call_arg

import (
	"context"

	"github.com/synthesio/zconfig/v2"
)

func detectNonPointerStruct() {
	var a struct{}
	zconfig.Configure(context.Background(), a) // want "argument used as configuration receiver is not a struct pointer"
}

func detectDoublePointer() {
	a := new(struct{})
	zconfig.Configure(context.Background(), &a) // want "argument used as configuration receiver is not a struct pointer"
}

func detectAnonymousStruct() {
	var a struct {
		Unresolved *bool `inject:"unresolved"`
	}
	zconfig.Configure(context.Background(), &a) // want "no source is provided for alias 'unresolved' used by target fields: Unresolved"
}

var anon struct {
	Unresolved *bool `inject:"unresolved"`
}

func detectAnonymousStruct2() {
	zconfig.Configure(context.Background(), &anon) // want "no source is provided for alias 'unresolved' used by target fields: Unresolved"
}

func detectAnonymousStruct3() {
	a := struct {
		Unresolved2 *int `inject:"unresolved2"`
	}{}
	zconfig.Configure(context.Background(), &a) // want "no source is provided for alias 'unresolved2' used by target fields: Unresolved2"
}

func detectAnonymousStruct4() {
	zconfig.Configure(context.Background(), new(struct { // want "no source is provided for alias 'unresolved3' used by target fields: Unresolved3"
		Unresolved3 *int `inject:"unresolved3"`
	}))
}
