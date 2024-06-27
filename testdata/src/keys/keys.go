package keys

type A struct { // want A:"<init:none>"
	B bool `key:"b"`
}

type EmbeddedNoKey struct { // want EmbeddedNoKey:"<init:none>"
	A      // want "key 'b' defined by field 'A.B' is already used by field 'B'"
	B bool `key:"b"` // want "key 'b' defined by field 'B' is already used by field 'A.B'"
}

type EmbeddedGenericNoKey struct { // want EmbeddedGenericNoKey:"<init:none>"
	Generic[bool]      // want "key 'field' defined by field 'Generic.Field' is already used by field 'Field'"
	Field         bool `key:"field"` // want "key 'field' defined by field 'Field' is already used by field 'Generic.Field'"
}

type EmbeddedWithKey struct { // want EmbeddedWithKey:"<init:none>"
	A `key:"a"`
	// No issues expected, A.B will have the key a.b
	B bool `key:"b"`
}

type FieldWithNoKey struct { // want FieldWithNoKey:"<init:none>"
	A A             // want "field A contains key tags but is not tagged with a key"
	B Generic[bool] // want "field B contains key tags but is not tagged with a key"
}

type Ambiguous struct { // want Ambiguous:"<init:none>"
	Field1 bool `key:"Field"` // want "key 'Field' used by field 'Field1' and key 'field' used by field Field2 have the same env format 'FIELD'"
	Field2 bool `key:"field"` // want "key 'field' used by field 'Field2' and key 'Field' used by field Field1 have the same env format 'FIELD'"

	A  A    `key:"a"`   // want "key 'a.b' used by field 'A.B' and key 'a_b' used by field AB have the same env format 'A_B'"
	AB bool `key:"a_b"` // want "key 'a_b' used by field 'AB' and key 'a.b' used by field A.B have the same env format 'A_B'"
}

type KeyDefinedInGeneric struct { // want KeyDefinedInGeneric:"<init:none>"
	Generic[struct { // want "key 'field.b' used by field 'Generic.Field.B' and key 'field_b' used by field B have the same env format 'FIELD_B'"
		B bool `key:"b"`
	}]
	B bool `key:"field_b"` // want "key 'field_b' used by field 'B' and key 'field.b' used by field Generic.Field.B have the same env format 'FIELD_B'"
}

type Generic[T any] struct { // want Generic:"<init:none>"
	Field T `key:"field"`
}
