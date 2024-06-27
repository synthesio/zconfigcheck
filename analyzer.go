package zconfigcheck

import (
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/static"
)

const (
	zconfigPkgName = "github.com/synthesio/zconfig"
	LinterName     = "zconfigcheck"
)

var Analyzer = &analysis.Analyzer{
	Name:      LinterName,
	Doc:       "zconfigcheck detects common zconfig issues",
	Run:       run,
	Requires:  []*analysis.Analyzer{inspect.Analyzer, buildssa.Analyzer},
	FactTypes: []analysis.Fact{new(wrapperFact), new(structFact), new(hasWrappersFact)},
}

func run(pass *analysis.Pass) (interface{}, error) {
	c := checker{
		Pass:       pass,
		Inspector:  pass.ResultOf[inspect.Analyzer].(*inspector.Inspector),
		SSA:        pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA),
		PkgStructs: make(map[types.Type]StructInfo),
	}

	// Scan all package structs and their dependencies
	c.checkStructs()

	// Detect all redundant calls to Init methods
	c.lookupInitCalls()

	// Find calls to zconfig to detect which structs are used as configurable root
	c.detectCalls()

	return nil, nil
}

type checker struct {
	Pass       *analysis.Pass
	Inspector  *inspector.Inspector
	SSA        *buildssa.SSA
	PkgStructs map[types.Type]StructInfo

	// callGraph must only be accessed via the CallGraph method
	callGraph *callgraph.Graph
}

func (c *checker) CallGraph() *callgraph.Graph {
	if c.callGraph != nil {
		return c.callGraph
	}

	c.callGraph = static.CallGraph(c.SSA.Pkg.Prog)
	return c.callGraph
}

// Issues is a collection of detected issues grouped by their position in the source code
type Issues map[token.Pos][]string

// Add adds one or more issues for a given source code position
func (i Issues) Add(pos token.Pos, issues ...string) {
	i[pos] = append(i[pos], issues...)
}

// Merge merges the receiver with the argument, returning a new Issues which
// contains the sum of the two.
// Neither the receiver nor the argument are modified.
func (i Issues) Merge(o Issues) Issues {
	merged := make(Issues)
	for pos, issues := range i {
		merged.Add(pos, issues...)
	}
	for pos, issues := range o {
		merged.Add(pos, issues...)
	}
	return merged
}

// Report is a helper to simplify reporting all contained issues
func (i Issues) Report(pass *analysis.Pass) {
	for pos, issues := range i {
		for _, issue := range issues {
			pass.Reportf(pos, issue)
		}
	}
}
