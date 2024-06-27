package init

import "context"

type WrongSignature struct{} // want WrongSignature:"<init:none>"
func (WrongSignature) Init() {}

type WrongSignature2 struct{}                    // want WrongSignature2:"<init:none>"
func (WrongSignature2) Init(ctx context.Context) {}

type WrongSignature3 struct{} // want WrongSignature3:"<init:none>"
func (WrongSignature3) Init(ctx context.Context) (bool, error) {
	return false, nil
}
