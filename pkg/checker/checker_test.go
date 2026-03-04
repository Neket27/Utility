package checker

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
	"utility/pkg/analyzer/rules"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

type mockRule struct {
	result *rules.RuleResult
}

func (m *mockRule) SetAutoFixEnabled(b bool) {
	panic("implement me")
}

func (m *mockRule) IsAutoFixEnabled() bool {
	panic("implement me")
}

func (m *mockRule) Check(ctx *rules.CheckContext) *rules.RuleResult {
	return m.result
}

func (m *mockRule) Name() string {
	return "mock"
}

func (m *mockRule) Description() string {
	return "mock rule"
}

func (m *mockRule) Enabled() bool {
	return true
}

func (m *mockRule) SetEnabled(enabled bool) {
}

func (m *mockRule) Configure(data map[string]any) error {
	return nil
}

type mockDetector struct {
	isMethod bool
	isLogger bool
}

func (m *mockDetector) IsLogMethod(name string) bool {
	return m.isMethod
}

func (m *mockDetector) IsLogger(pass *analysis.Pass, expr ast.Expr) bool {
	return m.isLogger
}

func newPass(t *testing.T, src string) (*analysis.Pass, *inspector.Inspector) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	insp := inspector.New([]*ast.File{file})

	pass := &analysis.Pass{
		Fset: fset,
		Files: []*ast.File{
			file,
		},
		ResultOf: map[*analysis.Analyzer]interface{}{
			inspect.Analyzer: insp,
		},
	}

	return pass, insp
}

func TestNew(t *testing.T) {
	r := &mockRule{}
	c := New([]rules.Rule{r})

	if c == nil {
		t.Fatal("expected checker to be created")
	}

	if len(c.rules) != 1 {
		t.Fatal("rules not set correctly")
	}

	if c.detector == nil {
		t.Fatal("detector must not be nil")
	}
}

func TestCheck_EarlyReturn_NotSelector(t *testing.T) {
	src := `
package test
func x() {
	println("hello")
}
`
	pass, _ := newPass(t, src)

	c := &Checker{
		rules:    []rules.Rule{},
		detector: &mockDetector{},
	}

	c.Check(pass)
}

func TestCheck_MethodNotLogger(t *testing.T) {
	src := `
package test
func x() {
	obj.NotLog("hello")
}
`
	pass, _ := newPass(t, src)

	c := &Checker{
		rules: []rules.Rule{},
		detector: &mockDetector{
			isMethod: false,
		},
	}

	c.Check(pass)
}

func TestCheck_NotLoggerInstance(t *testing.T) {
	src := `
package test
func x() {
	log.Info("hello")
}
`
	pass, _ := newPass(t, src)

	c := &Checker{
		rules: []rules.Rule{},
		detector: &mockDetector{
			isMethod: true,
			isLogger: false,
		},
	}

	c.Check(pass)
}

func TestCheck_RuleViolation(t *testing.T) {
	src := `
package test
func x() {
	log.Info("HELLO")
}
`
	pass, _ := newPass(t, src)

	violated := false

	pass.Report = func(d analysis.Diagnostic) {
		violated = true
	}

	rule := &mockRule{
		result: &rules.RuleResult{
			Passed:  false,
			Message: "error",
		},
	}

	c := &Checker{
		rules: []rules.Rule{rule},
		detector: &mockDetector{
			isMethod: true,
			isLogger: true,
		},
	}

	c.Check(pass)

	if !violated {
		t.Fatal("expected violation to be reported")
	}
}

func TestExecuteRules_NoArgs(t *testing.T) {
	call := &ast.CallExpr{}

	c := &Checker{}
	c.executeRules(call, &analysis.Pass{})
}

func TestReportViolation_WithSuggestedFix_StringLiteral(t *testing.T) {
	expr := &ast.BasicLit{
		Value: `"HELLO"`,
	}

	reported := false

	pass := &analysis.Pass{
		Report: func(d analysis.Diagnostic) {
			reported = true

			if len(d.SuggestedFixes) == 0 {
				t.Fatal("expected suggested fix")
			}
		},
	}

	c := &Checker{}

	result := &rules.RuleResult{
		Passed:  false,
		Message: "msg",
		SuggestedFix: &rules.SuggestedFix{
			Message: "fix",
			NewText: "hello",
		},
	}

	c.reportViolation(pass, expr, result)

	if !reported {
		t.Fatal("expected diagnostic")
	}
}

func TestExtractStringValue_BasicLiteral(t *testing.T) {
	expr := &ast.BasicLit{
		Kind: token.STRING, Value: `"hello"`,
	}

	val := extractStringValue(expr)

	if val != "hello" {
		t.Fatalf("unexpected value: %s", val)
	}
}

func TestExtractStringValue_Nil(t *testing.T) {
	if extractStringValue(nil) != "" {
		t.Fatal("expected empty string")
	}
}

func TestExtractStringValue_BinaryExpr(t *testing.T) {
	expr := &ast.BinaryExpr{
		X: &ast.BasicLit{Kind: token.STRING, Value: `"hello "`},
		Y: &ast.BasicLit{Kind: token.STRING, Value: `"world"`},
	}

	val := extractStringValue(expr)

	if val != "hello world" {
		t.Fatalf("unexpected value: %s", val)
	}
}

func TestExtractConcatenatedValue_Nested(t *testing.T) {
	expr := &ast.BinaryExpr{
		X: &ast.BinaryExpr{
			X: &ast.BasicLit{Kind: token.STRING, Value: `"a"`},
			Y: &ast.BasicLit{Kind: token.STRING, Value: `"b"`},
		},
		Y: &ast.BasicLit{Kind: token.STRING, Value: `"c"`},
	}

	val := extractConcatenatedValue(expr)

	if val != "abc" {
		t.Fatalf("unexpected: %s", val)
	}
}

func TestIsStringLiteral(t *testing.T) {
	expr := &ast.BasicLit{}

	if !isStringLiteral(expr) {
		t.Fatal("expected true")
	}

	if isStringLiteral(&ast.BinaryExpr{}) {
		t.Fatal("expected false")
	}
}
