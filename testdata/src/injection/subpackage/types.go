package subpackage

type InjectionSrc struct {
	Source *string `inject-as:"source"`
}

type InjectionTarget struct {
	Target *string `inject:"source"`
}

type WrongInjectionTarget struct {
	Target *int `inject:"source"`
}

type Doer interface {
	Do()
}

type DoerImpl struct{}

func (*DoerImpl) Do() {}
