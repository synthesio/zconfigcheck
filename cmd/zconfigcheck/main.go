package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"
	"zconfigcheck"
)

func main() {
	singlechecker.Main(zconfigcheck.Analyzer)
}
