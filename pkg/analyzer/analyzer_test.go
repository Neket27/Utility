package analyzer_test

import (
	"testing"
	"utility/pkg/analyzer"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	newAnalyzer := analyzer.NewAnalyzer()

	analysistest.Run(t, testdata, newAnalyzer, "example/example.go")
}
