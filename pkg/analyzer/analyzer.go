package analyzer

import (
	"fmt"
	"utility/pkg/analyzer/rules"
	"utility/pkg/analyzer/rules/config"
	"utility/pkg/checker"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

const Doc = `loglinter checks log messages for compliance with logging best practices.`

var configPath string

func NewAnalyzer() *analysis.Analyzer {
	analyzer := &analysis.Analyzer{
		Name:     "loglinter",
		Doc:      Doc,
		Run:      run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}

	analyzer.Flags.StringVar(&configPath, "config", "", "Path to config file")

	return analyzer
}

func run(pass *analysis.Pass) (interface{}, error) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	cfg.Print()

	rulesList := loadRules(cfg)

	checkerInstance := checker.New(rulesList)
	checkerInstance.Check(pass)

	return nil, nil
}

func loadConfig(path string) (*config.Config, error) {
	loader := config.NewLoader(path)
	return loader.Load()
}

func loadRules(cfg *config.Config) []rules.Rule {
	allRules, err := rules.RuleSet()
	if err != nil {
		return nil
	}
	enabledRules := make([]rules.Rule, 0, len(allRules))

	for _, rule := range allRules {
		var isEnabled bool
		var ruleConfig map[string]any

		switch rule.Name() {
		case "lowercase":
			if rc := cfg.Rules.Lowercase; rc != nil {
				isEnabled = rc.IsEnabled()
				ruleConfig = map[string]any{
					"enabled":          rc.IsEnabled(),
					"auto_fix_enabled": rc.IsAutoFixEnabled(),
				}
			} else {
				isEnabled = true
			}

		case "english_only":
			if rc := cfg.Rules.EnglishOnly; rc != nil {
				isEnabled = rc.IsEnabled()
			} else {
				isEnabled = true
			}

		case "no_special_chars":
			if ns := cfg.Rules.NoSpecialChars; ns != nil {
				isEnabled = ns.IsEnabled()
				ruleConfig = map[string]any{
					"enabled":              ns.IsEnabled(),
					"auto_fix_enabled":     ns.IsAutoFixEnabled(),
					"max_consecutive_dots": *ns.MaxConsecutiveDots,
				}
			} else {
				isEnabled = true
			}

		case "sensitive_words":
			if sw := cfg.Rules.SensitiveWords; sw != nil {
				isEnabled = sw.IsEnabled()
				ruleConfig = map[string]any{
					"enabled":          sw.IsEnabled(),
					"auto_fix_enabled": sw.IsAutoFixEnabled(),
				}
				if len(sw.Words) > 0 {
					ruleConfig["words"] = sw.Words
				}
				if len(sw.SafePhrases) > 0 {
					ruleConfig["safe_phrases"] = sw.SafePhrases
				}
			} else {
				isEnabled = true
			}

		case "custom_patterns":
			if cp := cfg.Rules.CustomPatterns; cp != nil {
				isEnabled = cp.IsEnabled()
				ruleConfig = map[string]any{
					"enabled": cp.IsEnabled(),
				}
				if len(cp.Patterns) > 0 {
					ruleConfig["patterns"] = cp.Patterns
				}
			} else {
				isEnabled = true
			}

		default:
			isEnabled = true
		}

		if !isEnabled {
			continue
		}
		if ruleConfig != nil {
			if err := rule.Configure(ruleConfig); err != nil {
				fmt.Printf("Warning: Configure error for %s: %v\n", rule.Name(), err)
			}
		}

		enabledRules = append(enabledRules, rule)
	}

	return enabledRules
}
