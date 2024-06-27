# zConfigCheck

`zconfigcheck` is a linter for [zconfig](https://github.com/synthesio/zconfig).
It detects a wide range of common mistakes which can lead to unexpected behavior.

This tool can either be used as a `go vet` tool or as a `golangci-lint` plugin.
See the [dedicated README.md file](golangci/README.md) for more information about using `zconfigcheck` with
`golangci-lint`.

## Installation

```console
$ go install github.com/synthesio/zconfigcheck@latest
```

## Usage

```console
$ go vet -vettool="$(which zconfigcheck)" TARGET_PKG
```

## Limitations

### Calls detection

`zconfigcheck` is only able to detect static calls to `zconfig`.
If calls to `zconfig` made by your code cannot be computed using a static call graph,
then some warnings will not be output.

### Argument parsing

`zconfigcheck` cannot tell whether a given type will be supported by
the `zconfig` default or custom parsers.
