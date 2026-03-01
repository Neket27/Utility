package rules

import (
	"fmt"
	"regexp"
)

const RuleCustomPatternsName = "custom_patterns"

type CustomPatternsRule struct {
	BaseRule
	compiledRegex []*regexp.Regexp
}

func NewCustomPatternsRule() Rule {
	return &CustomPatternsRule{
		BaseRule:      NewBaseRule(RuleCustomPatternsName, "Checks log messages against custom regex patterns"),
		compiledRegex: []*regexp.Regexp{},
	}
}

func (r *CustomPatternsRule) Configure(config map[string]any) error {
	if err := r.BaseRule.Configure(config); err != nil {
		return err
	}

	patterns, ok := config["patterns"].([]any)
	if !ok || len(patterns) == 0 {
		return nil
	}

	r.compiledRegex = make([]*regexp.Regexp, 0, len(patterns))
	for i, p := range patterns {
		s, ok := p.(string)
		if !ok {
			return fmt.Errorf("pattern at index %d is not a string", i)
		}

		re, err := regexp.Compile(s)
		if err != nil {
			return fmt.Errorf("compile pattern %q: %w", s, err)
		}

		r.compiledRegex = append(r.compiledRegex, re)
	}

	return nil
}

func (r *CustomPatternsRule) Check(ctx *CheckContext) *RuleResult {
	if !r.Enabled() || len(r.compiledRegex) == 0 {
		return ResultPass()
	}

	for _, re := range r.compiledRegex {
		if match := re.FindString(ctx.Msg); match != "" {
			return ResultFailWithSuggestion(
				fmt.Sprintf("log message matches forbidden pattern: %s", re.String()),
				"Remove or modify matching content",
				re.ReplaceAllString(ctx.Msg, "[REDACTED]"),
			)
		}
	}

	return ResultPass()
}
