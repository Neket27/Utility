package rules

import (
	"fmt"
	"os"
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

	patternsRaw, ok := config["patterns"]
	if !ok {
		fmt.Fprintf(os.Stderr, "[CustomPatterns] ERROR: patterns not found in config\n")
		return nil
	}

	fmt.Fprintf(os.Stderr, "[CustomPatterns] Raw patterns: %v (type: %T)\n", patternsRaw, patternsRaw)

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
	}

	fmt.Fprintf(os.Stderr, "[CustomPatterns] Parsed patterns: %v\n", patterns)

	if len(patterns) == 0 {
		fmt.Fprintf(os.Stderr, "[CustomPatterns] ERROR: no valid patterns\n")
		return nil
	}

	r.compiledRegex = make([]*regexp.Regexp, 0, len(patterns))
	for i, s := range patterns {
		fmt.Fprintf(os.Stderr, "[CustomPatterns] Compiling pattern %d: %q\n", i, s)

		re, err := regexp.Compile(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[CustomPatterns] ERROR compiling %q: %v\n", s, err)
			return fmt.Errorf("compile pattern %q: %w", s, err)
		}

		r.compiledRegex = append(r.compiledRegex, re)
		fmt.Fprintf(os.Stderr, "[CustomPatterns] Compiled regex %d: %q\n", i, re.String())
	}

	fmt.Fprintf(os.Stderr, "[CustomPatterns] Total compiled regexes: %d\n", len(r.compiledRegex))
	return nil
}

func (r *CustomPatternsRule) Check(ctx *CheckContext) *RuleResult {
	if !r.Enabled() {
		return ResultPass()
	}

	if len(r.compiledRegex) == 0 {
		fmt.Fprintf(os.Stderr, "[CustomPatterns] WARNING: no compiled regexes, skipping check\n")
		return ResultPass()
	}

	fmt.Fprintf(os.Stderr, "[CustomPatterns] Checking message: %q\n", ctx.Msg)

	for i, re := range r.compiledRegex {
		fmt.Fprintf(os.Stderr, "[CustomPatterns] Testing regex %d: %s\n", i, re.String())

		if match := re.FindString(ctx.Msg); match != "" {
			fmt.Fprintf(os.Stderr, "[CustomPatterns] ✓ MATCH! pattern=%s, match=%s\n", re.String(), match)
			return ResultFailWithSuggestion(
				fmt.Sprintf("log message matches forbidden pattern: %s", re.String()),
				"Remove or modify matching content",
				re.ReplaceAllString(ctx.Msg, "[REDACTED]"),
			)
		} else {
			fmt.Fprintf(os.Stderr, "[CustomPatterns] ✗ No match for pattern: %s\n", re.String())
		}
	}

	return ResultPass()
}
