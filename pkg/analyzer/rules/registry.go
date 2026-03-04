package rules

func RuleSet() ([]Rule, error) {
	builders := []RuleBuilder{
		NewSensitiveWordsRule,
		NewNoSpecialCharsRule,
		NewLowercaseRule,
		NewEnglishOnlyRule,
		NewCustomPatternsRule,
	}

	rules := make([]Rule, 0, len(builders))

	for _, builder := range builders {
		rules = append(rules, builder())
	}

	return rules, nil
}
