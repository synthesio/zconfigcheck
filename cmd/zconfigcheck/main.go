package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/synthesio/zconfigcheck"
)

func main() {
	singlechecker.Main(zconfigcheck.Analyzer)
}
