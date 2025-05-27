// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package static computes the call graph of a Go program containing
// only static call edges.
package static

import (
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

// CallGraph computes the call graph of the specified program
// considering only static calls.
func CallGraph(prog *ssa.Program) *callgraph.Graph {
	cg := callgraph.New(nil)

	for f := range ssautil.AllFunctions(prog) {
		fnode := cg.CreateNode(f)
		for _, b := range f.Blocks {
			for _, instr := range b.Instrs {
				if site, ok := instr.(ssa.CallInstruction); ok {
					if g := site.Common().StaticCallee(); g != nil {
						gnode := cg.CreateNode(g)
						callgraph.AddEdge(fnode, site, gnode)
					}
				}
			}
		}
	}

	return cg
}
