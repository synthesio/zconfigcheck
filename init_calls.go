package zconfigcheck

import (
	"go/token"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/ssa"
)

// walkGraph implements a postorder traversal of the callgraph
func walkGraph(node *callgraph.Node, fn func(*callgraph.Edge)) {
	seen := make(map[*callgraph.Node]bool)
	var visit func(n *callgraph.Node)
	visit = func(n *callgraph.Node) {
		if seen[n] {
			return
		}
		seen[n] = true
		for _, e := range n.Out {
			visit(e.Callee)
			fn(e)
		}
	}
	visit(node)
}

func getBaseValue(arg ssa.Value) ssa.Value {
	for {
		unOp, ok := arg.(*ssa.UnOp)
		if !ok {
			return arg
		}

		arg = unOp.X
	}
}

func getFieldAddr(value ssa.Value) (*ssa.FieldAddr, bool) {
	field, ok := getBaseValue(value).(*ssa.FieldAddr)
	if !ok {
		return nil, false
	}

	for {
		f, ok := getBaseValue(field.X).(*ssa.FieldAddr)
		if !ok {
			return field, true
		}
		field = f
	}
}

// lookupInitCalls detects and reports any redundant calls made to Init methods that are already called
// by zconfig.
// To avoid any false positive issues, only calls to Init done by an Init method are reported.
// This method supposes that checkStructs has been called on the same checker before, because it needs to
// access the complete PkgStructs map.
func (c *checker) lookupInitCalls() {
	inits := make(map[token.Pos]map[token.Pos]map[int]struct{})

	for _, info := range c.PkgStructs {
		if !info.HasInitMethod {
			// We only want to check calls done in the body of an Init method (or by one the functions it transitively calls)
			continue
		}

		// Store the position of the field Init method implementation (it is used to uniquely identify it) and the
		// position of the field it belongs to in the parent struct
		childrenInits := make(map[token.Pos]map[int]struct{})
		for _, child := range info.Children {
			if child.InitPos != token.NoPos && child.CallCount > 0 {
				if childrenInits[child.InitPos] == nil {
					childrenInits[child.InitPos] = make(map[int]struct{})
				}
				childrenInits[child.InitPos][child.Index] = struct{}{}
			}
		}

		inits[info.InitPos] = childrenInits
	}

	if len(inits) == 0 {
		// this package has no structs implementing the Init method, nothing to check
		return
	}

	graph := c.CallGraph()
	for _, node := range graph.Nodes {
		if node.Func == nil {
			continue
		}

		childrenInits, ok := inits[node.Func.Pos()]
		if !ok {
			// This node is not an Init method or its struct receiver does not have any other fields implementing an Init method
			continue
		}

		walkGraph(node, func(e *callgraph.Edge) {
			calleePos := e.Callee.Func.Pos()
			if calleePos == node.Func.Pos() {
				// this is a recursive call

				// do not report issues on methods generated due to generic type instantiation
				if e.Site.Pos().IsValid() {
					c.Pass.Reportf(e.Site.Pos(), "Init method is already invoked by zconfig")
				}

				return
			}

			indices, ok := childrenInits[calleePos]
			if !ok {
				return
			}

			args := e.Site.Common().Args
			if len(args) == 0 {
				// the called function has 0 arguments, so it cannot call any Init method
				return
			}

			field, ok := getFieldAddr(args[0])
			if !ok {
				return
			}

			if _, ok := indices[field.Field]; ok {
				c.Pass.Reportf(e.Site.Pos(), "Init method is already invoked by zconfig")
			}
		})
	}
}
