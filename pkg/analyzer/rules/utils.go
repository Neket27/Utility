package rules

import (
	"strings"
)

func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func ExtractSuggestedText(fix *SuggestedFix) string {
	if fix == nil {
		return ""
	}
	return fix.NewText
}
