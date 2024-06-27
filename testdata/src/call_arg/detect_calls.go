package call_arg

import (
	"context"

	"github.com/synthesio/zconfig/v2"
	"testdata/src/call_arg/subpackage"
)

func directConfigureCall() {
	zconfig.Configure(context.Background(), true) // want "argument used as configuration receiver is not a struct pointer"
}

func directProcessCall() {
	zconfig.DefaultProcessor.Process(context.Background(), true) // want "argument used as configuration receiver is not a struct pointer"
}

func configureWrapperCall() {
	subpackage.ConfigureWrapper(true) // want "argument used as configuration receiver is not a struct pointer"
}

func configureCtxWrapperCall() {
	subpackage.ConfigureCtxWrapper(context.Background(), true) // want "argument used as configuration receiver is not a struct pointer"
}

func processorWrapperCall() {
	subpackage.ProcessorWrapper(context.Background(), true) // want "argument used as configuration receiver is not a struct pointer"
}

func structMethodCall() {
	h := subpackage.ConfigHelper{}
	h.Configure(context.Background(), true) // want "argument used as configuration receiver is not a struct pointer"
}

func structMethodCall2() {
	h := subpackage.ConfigHelper{}
	h.DeferredConfigure(true) // want "argument used as configuration receiver is not a struct pointer"
}

func deferredCall() {
	defer func() {
		subpackage.ConfigureWrapper(true) // want "argument used as configuration receiver is not a struct pointer"
	}()
}

func deferredCall2() {
	subpackage.DeferredConfigureWrapper(true) // want "argument used as configuration receiver is not a struct pointer"
}

func deferredCall3() {
	h := subpackage.ConfigHelper{}
	h.DeferredConfigure(true) // want "argument used as configuration receiver is not a struct pointer"
}

func closureCall() {
	a := func(b any) {
		subpackage.ClosureConfigure(b)
	}
	a(true) // want "argument used as configuration receiver is not a struct pointer"
}

func anonymousFuncCall() func() {
	anon := func() {
		subpackage.ConfigureWrapper(true) // want "argument used as configuration receiver is not a struct pointer"
	}
	return anon
}

func anonymousFuncCallWithParam() {
	anon := func(s any) {
		subpackage.ConfigureWrapper(s)
	}

	anon(true) // want "argument used as configuration receiver is not a struct pointer"
}

func funcStoredInVar() {
	config := subpackage.ConfigureWrapper
	config(true) // want "argument used as configuration receiver is not a struct pointer"
}

func methodStoredInVar() {
	h := subpackage.ConfigHelper{}
	config := h.Configure
	config(context.Background(), true) // want "argument used as configuration receiver is not a struct pointer"
}

func goroutineCall() {
	go func() {
		subpackage.ConfigureWrapper(true) // want "argument used as configuration receiver is not a struct pointer"
	}()
}
