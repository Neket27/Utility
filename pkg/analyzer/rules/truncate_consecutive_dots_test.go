package rules

import "testing"

func TestTruncateConsecutiveDots(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxDots  int
		expected string
	}{
		{"maxDots=0: single dot", "server started.", 0, "server started"},
		{"maxDots=0: ellipsis", "waiting...", 0, "waiting"},
		{"maxDots=0: multiple groups", "a. b.. c...", 0, "a b c"},
		{"maxDots=1: keep single", "server started.", 1, "server started."},
		{"maxDots=1: truncate ellipsis", "waiting...", 1, "waiting."},
		{"maxDots=2: keep two", "waiting..", 2, "waiting.."},
		{"maxDots=2: truncate three", "waiting...", 2, "waiting.."},
		{"dots reset after space", "a. b. c...", 1, "a. b. c."},
		{"dots reset after letter", "a.b.c", 1, "a.b.c"},
		{"no dots", "hello world", 1, "hello world"},
		{"empty string", "", 1, ""},
		{"only dots", "...", 1, "."},
		{"negative maxDots", "test...", -1, "test..."}, // возвращает как есть
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateConsecutiveDots(tt.input, tt.maxDots)
			if got != tt.expected {
				t.Errorf("truncateConsecutiveDots(%q, %d) = %q, want %q",
					tt.input, tt.maxDots, got, tt.expected)
			}
		})
	}
}
