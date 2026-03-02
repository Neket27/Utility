package loglinter

import (
	"golang.org/x/tools/go/analysis"
	"utility/pkg/analyzer"
)

type plugin struct{}

func (p *plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		analyzer.NewAnalyzer(),
	}, nil
}

func (p *plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

func New(settings any) (register.LinterPlugin, error) {
	return &plugin{}, nil
}

func init() {
	register.Plugin("loglinter", New)
}
