package rules

import (
	"go/ast"
)

type RuleResult struct {
	Passed       bool
	Message      string
	SuggestedFix *SuggestedFix
}

type SuggestedFix struct {
	Message string
	NewText string
}

type CheckContext struct {
	MsgExpr ast.Expr
	Msg     string
}

type Rule interface {
	Name() string
	Description() string
	Enabled() bool
	SetEnabled(enabled bool)
	SetAutoFixEnabled(b bool)
	Configure(config map[string]any) error
	Check(ctx *CheckContext) *RuleResult
	IsAutoFixEnabled() bool
}

type RuleBuilder func() Rule

type BaseRule struct {
	name           string
	description    string
	enabled        bool
	autoFixEnabled bool
}

func NewBaseRule(name, description string) BaseRule {
	return BaseRule{
		name:           name,
		description:    description,
		enabled:        true,
		autoFixEnabled: true,
	}
}

func (b *BaseRule) Name() string             { return b.name }
func (b *BaseRule) Description() string      { return b.description }
func (b *BaseRule) Enabled() bool            { return b.enabled }
func (b *BaseRule) SetEnabled(enabled bool)  { b.enabled = enabled }
func (b *BaseRule) AutoFixEnabled() bool     { return b.autoFixEnabled }
func (b *BaseRule) SetAutoFixEnabled(v bool) { b.autoFixEnabled = v }

func (b *BaseRule) Configure(config map[string]any) error {
	if enabled, ok := config["enabled"].(bool); ok {
		b.enabled = enabled
	}

	if autoFix, ok := config["auto_fix_enabled"].(bool); ok {
		b.autoFixEnabled = autoFix
	}
	return nil
}

func (b *BaseRule) IsAutoFixEnabled() bool {
	return b.AutoFixEnabled()
}

func ResultPass() *RuleResult {
	return &RuleResult{Passed: true}
}

func ResultFail(message string) *RuleResult {
	return &RuleResult{Passed: false, Message: message}
}

func ResultFailWithSuggestion(message, suggestionMessage, newText string) *RuleResult {
	return &RuleResult{
		Passed:  false,
		Message: message,
		SuggestedFix: &SuggestedFix{
			Message: suggestionMessage,
			NewText: newText,
		},
	}
}
