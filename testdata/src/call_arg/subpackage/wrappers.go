package subpackage // want package:"has wrappers"

import (
	"context"

	"github.com/synthesio/zconfig/v2"
)

func ConfigureCtxWrapper(ctx context.Context, s any) { // want ConfigureCtxWrapper:"wrapper, arg: 1"
	err := zconfig.Configure(ctx, s)
	if err != nil {
		panic(err)
	}
}

func ConfigureWrapper(s any) { // want ConfigureWrapper:"wrapper, arg: 0"
	ConfigureCtxWrapper(context.Background(), s)
}

func DeferredConfigureWrapper(s any) { // want DeferredConfigureWrapper:"wrapper, arg: 0"
	defer func() {
		ConfigureWrapper(s)
	}()
}

func ClosureConfigure(s any) { // want ClosureConfigure:"wrapper, arg: 0"
	a := func(r any) {
		ConfigureWrapper(r)
	}
	b := a
	b(s)
}

func ProcessorWrapper(ctx context.Context, s any) { // want ProcessorWrapper:"wrapper, arg: 1"
	err := zconfig.DefaultProcessor.Process(ctx, s)
	if err != nil {
		panic(err)
	}
}

type ConfigHelper struct { // want ConfigHelper:"<init:none>"
}

func (c ConfigHelper) Configure(ctx context.Context, s any) { // want Configure:"wrapper, arg: 2, is method"
	ProcessorWrapper(ctx, s)
}

func (c ConfigHelper) DeferredConfigure(s any) { // want DeferredConfigure:"wrapper, arg: 1, is method"
	defer func(a string) {
		defer func(b int) {
			c := s
			defer func(s any) {
				defer func() {
					DeferredConfigureWrapper(s)
				}()
			}(c)
		}(1)
	}("a")
}

func (c ConfigHelper) GoroutineConfigure(s any) { // want GoroutineConfigure:"wrapper, arg: 1, is method"
	go func(b any) {
		DeferredConfigureWrapper(b)
	}(s)
}
