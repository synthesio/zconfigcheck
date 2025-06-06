// Package foo is a package that validates zconfigcheck works correctly
package foo

import (
	"context"

	"github.com/synthesio/zconfig/v2"
)

func init() {
	var a struct{}
	zconfig.Configure(context.Background(), a)
}
