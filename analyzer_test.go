package zconfigcheck_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/synthesio/zconfigcheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get wd: %s", err)
	}
	testdata := filepath.Join(wd, "testdata")

	for testName, pkgName := range map[string]string{
		"call argument detection": "/call_arg",
		"call chains detection":   "call_arg/subpackage",
		"init method detection":   "init",
		"tag parsing":             "tags",
		"injection":               "injection",
		"keys":                    "keys",
		"dependency cycles":       "cycles",
	} {
		t.Run(testName, func(t *testing.T) {
			analysistest.Run(t, testdata, zconfigcheck.Analyzer, "testdata/src/"+pkgName)
		})
	}
}
