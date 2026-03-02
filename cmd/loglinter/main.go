package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"
	"utility/pkg/analyzer"
)

func main() {
	singlechecker.Main(analyzer.NewAnalyzer())
}
