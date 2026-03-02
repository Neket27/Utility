package rules

import (
	"fmt"
	"regexp"
)

const RuleCustomPatternsName = "custom_patterns"

type CustomPatternsRule struct {
	BaseRule
	CompiledRegex []*regexp.Regexp
}

func NewCustomPatternsRule() Rule {
	return &CustomPatternsRule{
		BaseRule:      NewBaseRule(RuleCustomPatternsName, "Checks log messages against custom regex patterns"),
		CompiledRegex: []*regexp.Regexp{},
	}
}

func (r *CustomPatternsRule) Configure(config map[string]any) error {
	if err := r.BaseRule.Configure(config); err != nil {
		return err
	}

	patternsRaw, ok := config["patterns"]

	if !ok {
		return nil
	}

	var patterns []string
	switch v := patternsRaw.(type) {
	case []any:
		for _, p := range v {
			if s, ok := p.(string); ok {
				patterns = append(patterns, s)
			}
		}
	case []string:
		patterns = v
	default:
		patterns = []string{}
	}

	if len(patterns) == 0 {
		r.CompiledRegex = []*regexp.Regexp{}
		return nil
	}

	r.CompiledRegex = make([]*regexp.Regexp, 0, len(patterns))
	for _, s := range patterns {
		re, err := regexp.Compile(s)
		if err != nil {
			return fmt.Errorf("compile pattern %q: %w", s, err)
		}
		r.CompiledRegex = append(r.CompiledRegex, re)
	}

	return nil
}

func (r *CustomPatternsRule) Check(ctx *CheckContext) *RuleResult {
	if !r.Enabled() || len(r.CompiledRegex) == 0 {
		return ResultPass()
	}

	for _, re := range r.CompiledRegex {
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
