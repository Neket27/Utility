package test

import (
	"strings"
	"utility/pkg/analyzer/rules"
)

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func extractSuggestedText(fix *rules.SuggestedFix) string {
	if fix == nil {
		return ""
	}
	return fix.NewText
}
