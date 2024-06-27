package zconfigcheck

import (
	"fmt"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/ssa"
)

// getArgIssues returns any issues arising when the given value is used as the argument for a call to zconfig.Configure
func (c *checker) getArgIssues(arg ssa.Value) []string {
	typ := getStructType(arg)
	if typ == nil {
		return []string{"argument used as configuration receiver is not a struct pointer"}
	}

	if info, ok := c.PkgStructs[typ]; ok {
		// this struct was declared in this package, so we can directly access its issues
		return info.Fact().Issues
	}

	named, ok := typ.(*types.Named)
	if !ok {
		// this should never happen, because the language does not allow importing anonymous structs
		// from other packages
		return []string{"unexpected anonymous struct"}
	}

	var fact structFact
	if c.Pass.ImportObjectFact(named.Obj(), &fact) {
		return fact.Issues
	}

	// this should never happen
	return []string{"cannot find any information about the struct"}
}

// getStructType returns the *types.Struct matching the given value.
// When the value is not a struct pointer, nil is returned.
func getStructType(arg ssa.Value) types.Type {
	typ := arg.Type().Underlying()
	if makeItf, ok := arg.(*ssa.MakeInterface); ok {
		typ = makeItf.X.Type().Underlying()
	}

	ptr, ok := typ.(*types.Pointer)
	if !ok {
		return nil
	}

	elem := ptr.Elem()
	if _, ok := elem.Underlying().(*types.Struct); ok {
		return elem
	}

	return nil
}

// scanPackage returns true if the current package might contain any calls to zconfig.
func (c *checker) scanPackage() bool {
	if strings.HasPrefix(c.Pass.Pkg.Path(), zconfigPkgName) {
		// the current package belongs to zconfig itself, we need to scan it
		return true
	}

	// build a dictionary containing all package paths containing transitive calls to zconfig
	packagesWithWrappers := make(map[string]struct{})
	for _, fact := range c.Pass.AllPackageFacts() {
		if _, ok := fact.Fact.(*hasWrappersFact); ok {
			packagesWithWrappers[fact.Package.Path()] = struct{}{}
		}
	}

	if len(packagesWithWrappers) == 0 {
		// If we didn't find any calls to zconfig yet, then the package cannot contain any calls to it.
		// This can only happen before the zconfig package is scanned. It helps skipping at least all internal
		// golang packages.
		return false
	}

	// We want to scan this package only if it imports zconfig or another package which wraps it.
	for _, importedPkg := range c.Pass.Pkg.Imports() {
		if _, ok := packagesWithWrappers[importedPkg.Path()]; ok {
			return true
		}
	}
	return false
}

// detectCalls builds and walks the current package static call graph to
// detect any calls made to zconfig/Processor.Process (this includes zconfig.Configure).
// Any functions or methods which do not use a local variable as the call
// argument but rather one of their signature parameters are identified
// as wrappers.
// Any call to one of these wrappers will then be considered as a call to
// zconfig.
func (c *checker) detectCalls() {
	if !c.scanPackage() {
		return
	}

	if strings.HasPrefix(c.Pass.Pkg.Path(), zconfigPkgName) {
		// This is the zconfig package, we need to find and store the object for
		// the zconfig/Processor.Process method
		processor := c.SSA.Pkg.Type("Processor")
		if processor == nil {
			return
		}

		fnObj, _, _ := types.LookupFieldOrMethod(processor.Type(), true, c.Pass.Pkg, "Process")

		c.Pass.ExportObjectFact(fnObj, &wrapperFact{IsInvoke: true, ArgIndex: 2})
		c.Pass.ExportPackageFact(new(hasWrappersFact))
	}

	wrappers := wrapperRepository{
		wrappers: make(map[*ssa.Function]wrapper),
		pass:     c.Pass,
	}

	graph := c.CallGraph()
	for _, node := range graph.Nodes {
		// Try to find a call path leading to one of the known wrappers
		path := callgraph.PathSearch(node, func(node *callgraph.Node) bool {
			_, ok := wrappers.get(node)
			return ok
		})
		if len(path) == 0 {
			continue
		}

		if _, ok := path[0].Site.(*ssa.Call); !ok {
			// If the path starts with anything other than a call, then it could
			// be an anonymous function call, a call to defer or a goroutine.
			// In these cases we know that there are longer paths which include this one
			// as their "suffix", so we prefer processing those paths.
			continue
		}

		// Walk the found path starting from the end.
		// For each path item, we already know the callee:
		// - for the last item it was found in the repository during the call to PathSearch
		// - for other items it was added to the repository by previous iterations
		for i := len(path) - 1; i >= 0; i-- {
			caller := path[i].Caller

			if _, ok := wrappers.get(caller); ok {
				continue
			}

			for _, edge := range caller.Out {
				// Scan all calls made by the caller, because PathSearch only returns one
				// path amongst all possible ones.
				// If the called function receives one of the callers parameter as its argument,
				// then the caller will be marked as a wrapper.
				arg := wrappers.scan(edge)
				if arg == nil {
					// The caller is a wrapper: the argument used for the call is one of its parameters,
					// so there is nothing to do.
					continue
				}

				// Scan the argument used for the call to check whether it has the right type and report
				// any eventual issues.
				for _, issue := range c.getArgIssues(arg) {
					c.Pass.Reportf(edge.Site.Common().Pos(), issue)
				}
			}
		}
	}

	if wrappers.exported {
		c.Pass.ExportPackageFact(new(hasWrappersFact))
	}
}

// wrapper contains all necessary information to trace the variable which
// it eventually uses as an argument for a call to another wrapper or to zconfig/Processor.Process
type wrapper struct {
	IsInvoke bool
	ArgIndex int
	FreeVar  *ssa.FreeVar
}

// Fact converts wrapper into a wrapperFact. wrapper does not implement
// the Fact interface because it has references to types which cannot be encoded by
// the gob package, leading to issues with golangci-lint cache.
func (w wrapper) Fact() wrapperFact {
	return wrapperFact{
		IsInvoke: w.IsInvoke,
		ArgIndex: w.ArgIndex,
	}
}

// Argument accepts as its parameter a call to a function or method which is represented by the receiver.
// For the given call, it returns the ssa.Value matching the variable used as the argument which will
// eventually be used by zconfig as the configuration root.
func (w wrapper) Argument(call *ssa.CallCommon) ssa.Value {
	if w.FreeVar != nil {
		return w.FreeVar
	}

	argIndex := w.ArgIndex
	if w.IsInvoke && argIndex >= len(call.Args) {
		argIndex--
	}
	arg := call.Args[argIndex]
	if unOp, ok := arg.(*ssa.UnOp); ok {
		arg = unOp.X
	}
	return arg
}

// wrapperFact contains a subset of wrapper information, in order to comply with
// golangci-lint gob encoding
type wrapperFact struct {
	IsInvoke bool
	ArgIndex int
}

func (wrapperFact) AFact() {}

func (w wrapperFact) String() string {
	s := fmt.Sprintf("wrapper, arg: %d", w.ArgIndex)
	if w.IsInvoke {
		s += ", is method"
	}
	return s
}

// hasWrappersFact is a Fact used to mark packages as containing wrapper
// functions (or methods). This allows to skip a costly static call graph build
// for packages which do not import any wrapper-containing packages.
type hasWrappersFact struct{}

func (hasWrappersFact) AFact() {}

func (hasWrappersFact) String() string {
	return "has wrappers"
}

// wrapperRepository is a repository which grants access to all detected wrappers.
// It must be instantiated or reset to its zero-value for each analyzed package.
type wrapperRepository struct {
	wrappers map[*ssa.Function]wrapper
	pass     *analysis.Pass
	exported bool
}

// get returns a wrapper and a boolean to indicate whether the given node
// argument is a known wrapper.
func (w *wrapperRepository) get(node *callgraph.Node) (wrapper, bool) {
	if node == nil {
		return wrapper{}, false
	}

	fn := node.Func
	if fn == nil {
		return wrapper{}, false
	}

	if wrapper, ok := w.wrappers[fn]; ok {
		// This wrapper belongs to the current package, no need to import a fact
		return wrapper, true
	}

	if fn.Object() == nil {
		return wrapper{}, false
	}

	var fact wrapperFact
	if !w.pass.ImportObjectFact(fn.Object(), &fact) {
		return wrapper{}, false
	}

	// Convert the wrapperFact back to the wrapper type
	return wrapper{
		IsInvoke: fact.IsInvoke,
		ArgIndex: fact.ArgIndex,
	}, true
}

// add adds the given node and its related wrapper information to the repository.
// The wrapper information is added to the local package-scoped storage and also
// exported as a fact if it is a public function or method.
func (w *wrapperRepository) add(node *callgraph.Node, wrapper wrapper) {
	fn := node.Func

	wrapper.IsInvoke = fn.Signature.Recv() != nil
	w.wrappers[fn] = wrapper

	if fn.Object() == nil {
		// The given node is private or anonymous, no need to export a fact
		return
	}

	fact := wrapper.Fact()
	w.pass.ExportObjectFact(fn.Object(), &fact)
	w.exported = true
}

// scan returns:
//   - a non-nil ssa.Value if the given edge represents a call to a wrapper which uses a local variable as its argument.
//     See wrapper.Argument for more details on the returned ssa.Value.
//   - nil if the given edge does not represent a call to a wrapper, or it represents an intermediate call,
//     meaning a call done by a wrapper to another wrapper. In the second case the wrapper information is also
//     added to the repository.
func (w *wrapperRepository) scan(edge *callgraph.Edge) ssa.Value {
	wrapperInfo, ok := w.get(edge.Callee)
	if !ok {
		// this should never happen
		return nil
	}

	caller := edge.Caller

	arg := wrapperInfo.Argument(edge.Site.Common())

	for i, param := range caller.Func.Params {
		if param.Pos() == arg.Pos() {
			w.add(caller, wrapper{
				ArgIndex: i,
			})
			return nil
		}
	}

	if freeVar, ok := arg.(*ssa.FreeVar); ok {
		w.add(caller, wrapper{
			FreeVar: freeVar,
		})
		return nil
	}

	return arg
}
