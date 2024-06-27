package golangci

import (
	"github.com/golangci/plugin-module-register/register"
	"github.com/synthesio/zconfigcheck"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin(zconfigcheck.LinterName, New)
}

func New(_ any) (register.LinterPlugin, error) {
	return &Plugin{}, nil
}

type Plugin struct{}

func (p *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{zconfigcheck.Analyzer}, nil
}

func (p *Plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
