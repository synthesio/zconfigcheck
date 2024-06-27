package cycles

import (
	"context"

	"github.com/synthesio/zconfig/v2"
	"testdata/src/cycles/subpackage"
)

type A struct { // want A:"<init:none>"
	B B
	D D
}

type B struct { // want B:"<init:none>"
	C C
	A *A
}

type C struct { // want C:"<init:none>"
	A *A
}

type D struct { // want D:"<init:none>"
	E *bool `inject:"e"`
}

var _ = zconfig.Configure(context.Background(), new(A)) /* want
"configured struct contains dependency cycle: testdata/src/cycles.A -> testdata/src/cycles.B -> testdata/src/cycles.C -> testdata/src/cycles.A"
"configured struct contains dependency cycle: testdata/src/cycles.A -> testdata/src/cycles.B -> testdata/src/cycles.A"
"no source is provided for alias 'e' used by target fields: D.E"
*/

type Generic[T any] struct { // want Generic:"<init:none>"
	Field T
}

type F struct { // want F:"<init:none>"
	*Generic[F]
}

var _ = zconfig.Configure(context.Background(), new(Generic[F])) /* want
"configured struct contains dependency cycle: testdata/src/cycles.Generic\\[testdata/src/cycles.F\\] -> testdata/src/cycles.F -> testdata/src/cycles.Generic\\[testdata/src/cycles.F\\]"
*/

type genericF Generic[F] // want genericF:"<init:none>"

var _ = zconfig.Configure(context.Background(), &genericF{}) /* want
"configured struct contains dependency cycle: testdata/src/cycles.genericF -> testdata/src/cycles.F -> testdata/src/cycles.Generic\\[testdata/src/cycles.F\\] -> testdata/src/cycles.F"
*/

type genericF2 = Generic[F] // want genericF2:"<init:none>"

var _ = zconfig.Configure(context.Background(), &genericF2{}) /* want
"configured struct contains dependency cycle: testdata/src/cycles.Generic\\[testdata/src/cycles.F\\] -> testdata/src/cycles.F -> testdata/src/cycles.Generic\\[testdata/src/cycles.F\\]"
*/

type G struct { // want G:"<init:none>"
	subpackage.Generic[*G]
}

var _ = zconfig.Configure(context.Background(), new(G)) /* want
"configured struct contains dependency cycle: testdata/src/cycles.G -> testdata/src/cycles/subpackage.Generic\\[\\*testdata/src/cycles.G\\] -> testdata/src/cycles.G"
*/
