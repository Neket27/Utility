package analyzer

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"utility/pkg/analyzer/rules"
	"utility/pkg/checker"
)

const Doc = `loglinter checks log messages for compliance with logging best practices.

Rules:
1. Message must start with a lowercase letter
2. Message must be in English only (ASCII)
3. No special characters or emojis
4. No sensitive data (passwords, tokens, etc.)`

func NewAnalyzer(cfg any) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "loglinter",
		Doc:      Doc,
		Run:      makeRunFunc(cfg),
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

func makeRunFunc(cfg any) func(*analysis.Pass) (interface{}, error) {
	config := parseConfig(cfg)
	rulesList := loadRules(config)
	checker := checker.New(rulesList)

	return func(pass *analysis.Pass) (interface{}, error) {
		checker.Check(pass)
		return nil, nil
	}
}

func loadRules(cfg rulesConfig) []rules.Rule {
	allRules, _ := rules.GetAllRules()
	enabledRules := make([]rules.Rule, 0, len(allRules))

	for _, rule := range allRules {
		enabled := true
		if rc, exists := cfg.Rules[rule.Name()]; exists {
			if rc.Enabled != nil {
				enabled = *rc.Enabled
			}
			if len(rc.Data) > 0 {
				rule.Configure(rc.Data)
			}
		}
		if enabled {
			enabledRules = append(enabledRules, rule)
		}
	}

	return enabledRules
}
