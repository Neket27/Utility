package analyzer_test

import (
	"testing"
	"utility/pkg/analyzer"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analyzer := analyzer.NewAnalyzer(testdata)

	analysistest.Run(t, testdata, analyzer, "example/example.go")
}
