#!/bin/bash
# This script installs a specific version of golangci-lint and verifies the installation.

set -euo pipefail

if [ ! "$#" -eq 1 ]; then
	echo "Usage: $0 [latest|<golangci_version>]"
	echo
	echo "Examples:"
	echo "$0 latest # this would fetch the latest golangci-lint version"
	echo "$0 v2.1.6"
	exit 1
fi

version="$1"
case "$version" in
latest)
	# this one is special case, we want to fetch the latest version
	# the version will be determined later
	version="latest"
	;;
v*)
	# Version already starts with 'v', do nothing
	;;
*)
	version="v${version#v}" # Prepend 'v' to the version
	;;
esac

golangci_testbin="./golangci-lint"
gcl_testbin="./custom-gcl"

cd "$(dirname "$0")"

# Disable color output for the scripts call
export NO_COLOR=1

# Disable color output for golangci-lint verbose output
# This was fixed in golangci-lint > v2.1.6, but we still need to set it for older versions
export CLICOLOR=0

echo "🚧 Installing a local golangci-lint version…"
if ! curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b . "$version"; then
	echo "😵 Failed to install golangci-lint version $version"
	exit 1
fi

if [ ! -f "$golangci_testbin" ]; then
	echo "golangci-lint binary not found after installation"
	exit 1
fi

echo
echo "✅ golangci-lint version $version has been successfully installed."
if ! "$golangci_testbin" version; then
	echo "😵 Failed to run golangci-lint version command"
	exit 1
fi

# This is the special case for the latest version
# where we need to determine the actual version
if [ "$version" = "latest" ]; then
	echo "🚧 Computing latest golangci-lint version…"
	version="$($golangci_testbin version --short)"
	version="v${version#v}" # Ensure version starts with 'v'
	echo "✅ Found version: $version"
fi

echo
echo "🚧 configuring golangci-lint $version with zconfigcheck"
sed "s/GOLANGCI_VERSION.*/$version/g" <./.custom-gcl.tpl.yml >./.custom-gcl.yml
echo "✅ golangci-lint configured with:"
cat ./.custom-gcl.yml

echo
echo "🚧 Building local custom golangci-lint with zconfigcheck module…"
if ! "$golangci_testbin" --color=never custom --verbose; then
	echo "😵 Failed to build custom golangci-lint"
	exit 1
fi

if [ ! -f "$gcl_testbin" ]; then
	echo "😵 Custom golangci-lint binary not found after build"
	exit 1
fi
echo
echo "✅ custom golangci-lint has been successfully compiled and installed."
"$gcl_testbin" version

echo
echo "🚧 Checking if custom golangci-lint has zconfigcheck…"
if ! "$gcl_testbin" linters --enable-only zconfigcheck >/dev/null; then
	# If the linter is not found, can dump the linters list, by running the command again
	"$gcl_testbin" linters --enable-only zconfigcheck || true
	echo "😵 zconfigcheck is not enabled in custom golangci-lint"
	exit 1
fi
echo "✅ zconfigcheck is enabled in custom golangci-lint."

echo
echo "🚧 Running custom golangci-lint with zconfigcheck enabled…"
echo "NOTE: we expect to report issue with foo.go file"
if "$gcl_testbin" run --enable-only zconfigcheck foo/foo.go; then
	echo "😵 No issues found, but expected some issues to be reported."
	exit 1
fi
echo "✅ Custom golangci-lint with zconfigcheck module has been successfully built and tested."
