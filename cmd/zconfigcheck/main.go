package main

import (
	"github.com/synthesio/zconfigcheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(zconfigcheck.Analyzer)
}
