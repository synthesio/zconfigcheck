package tags

type doer interface {
	Do()
}

type doerImpl struct{} // want doerImpl:"<init:none>"

func (doerImpl) Do() {}

type EmptyTags struct { // want EmptyTags:"<init:none>"
	EmptyKey      bool      `key:"" description:""` // want "key tag cannot be empty" "description tag cannot be empty"
	EmptyInject   *doerImpl `inject:""`             // want "inject tag cannot be empty"
	EmptyInjectAs *doerImpl `inject-as:""`          // want "inject-as tag cannot be empty"
}

type PrivateFields struct { // want PrivateFields:"<init:none>"
	privateKey      bool  `key:"private"`       // want "private fields cannot have tags"
	privateInject   *bool `inject:"private"`    // want "private fields cannot have tags"
	privateInjectAs *bool `inject-as:"private"` // want "private fields cannot have tags"
}

type AlreadyDefined struct { // want AlreadyDefined:"<init:none>"
	A  bool      `key:"a"`       // want "key 'a' defined by field 'A' is already used by field 'A2'"
	A2 bool      `key:"a"`       // want "key 'a' defined by field 'A2' is already used by field 'A'"
	B  *doerImpl `inject-as:"b"` // want "inject-as alias 'b' defined by field B is already used by field B2"
	B2 *doerImpl `inject-as:"b"` // want "inject-as alias 'b' defined by field B2 is already used by field B"
}

type TypeErrors struct { // want TypeErrors:"<init:none>"
	Inject *doerImpl `inject:"doer"` /* want
	"injection alias 'doer': target fields 'Inject \\*testdata/src/tags.doerImpl' and 'InjectNonPtr testdata/src/tags.doerImpl' are incompatible, mismatched types"
	"injection alias 'doer': target field 'Inject \\*testdata/src/tags.doerImpl' cannot be injected with source field 'InjectAsNonPtr testdata/src/tags.doerImpl', mismatched types"
	*/
	InjectNonPtr doerImpl `inject:"doer"` /* want
	"field type is not a pointer nor interface, cannot be used as injection target"
	"injection alias 'doer': target fields 'InjectNonPtr testdata/src/tags.doerImpl' and 'Inject \\*testdata/src/tags.doerImpl' are incompatible, mismatched types"
	*/
	InjectAsNonPtr doerImpl `inject-as:"doer"` /* want
	"field type is not a pointer, cannot be used as injection source"
	"injection alias 'doer': cannot inject source field 'InjectAsNonPtr testdata/src/tags.doerImpl' into target field 'Inject \\*testdata/src/tags.doerImpl', mismatched types"
	*/
	InjectAsItf doer `inject-as:"itf"` // want "field type is not a pointer, cannot be used as injection source"
}

type IncompatibleTags struct { // want IncompatibleTags:"<init:none>"
	A *bool `key:"a" inject:"a"`       // want "key and inject tags should not be used on the same field"
	B bool  `default:""`               // want "default tag is used on field without key tag"
	C bool  `description:"c"`          // want "description tag is used on field without key tag"
	D *bool `inject-as:"d" inject:"d"` // want "inject and inject-as tags cannot be used on the same field"
}

type AllOk struct { // want AllOk:"<init:none>"
	A bool `key:"a"`
	B *int `key:"b" inject-as:"b"`
	C *int `inject:"b"`
	D doer `inject:"doer"`
	// No expected errors here, using the same inject alias many times is legitimate as long as types are coherent
	D2 *doerImpl `inject:"doer"`
}
