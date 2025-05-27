# About this repository

## Context

The `callgraph.go` file in this directory contains code from an old version of golang.org/x/tools

https://github.com/golang/tools/blob/v0.24.0/go/callgraph/static/static.go

This method was useful for retrieving all nodes, including the wrappers.

The ability of getting the wrappers were removed with the following commit:

https://github.com/golang/tools/commit/c538e2c079ea0676702503ccfe54f8c3ddd7ae81

The commit was made as a performance improvement, but it removed the ability to get the wrapper.

zconfigcheck needs to be able to analyze wrapped code.

For this reason, we imported the old code as-is.

## Warning

For the record, the method we are using is planned for deprecation:

- https://github.com/golang/go/issues/69231
- https://github.com/golang/go/issues/69291
- https://github.com/golang/go/issues/66251
- https://github.com/golang/go/issues/65915

So this might not work forever, but for now, it's OK.

## License

Copyright 2009 The Go Authors.

See [./LICENSE] imported from [github.com/golang/tools](https://github.com/golang/tools/blob/HEAD/LICENSE)