package rules

import (
	"sync"
)

var (
	registeredRules = make(map[string]RuleBuilder)
	initOnce        sync.Once
)

func GetAllRules() ([]Rule, error) {
	initOnce.Do(func() {
		RegisterRule(RuleLowercaseName, NewLowercaseRule)
		RegisterRule(RuleNoSpecialCharsName, NewNoSpecialCharsRule)
		RegisterRule(RuleEnglishOnlyName, NewEnglishOnlyRule)
		RegisterRule(RuleSensitiveWordsName, NewSensitiveWordsRule)
		RegisterRule(RuleCustomPatternsName, NewCustomPatternsRule)
	})

	rules := make([]Rule, 0, len(registeredRules))
	for _, builder := range registeredRules {
		rules = append(rules, builder())
	}
	return rules, nil
}

func RegisterRule(name string, builder RuleBuilder) {
	registeredRules[name] = builder
}
