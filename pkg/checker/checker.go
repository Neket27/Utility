package checker

import (
	"go/ast"
	"strings"
	"utility/pkg/analyzer/rules"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

type Logger interface {
	IsLogMethod(name string) bool
	IsLogger(pass *analysis.Pass, expr ast.Expr) bool
}

type Checker struct {
	rules    []rules.Rule
	detector Logger
}

func New(rulesList []rules.Rule) *Checker {
	return &Checker{
		rules:    rulesList,
		detector: NewLoggerDetector(),
	}
}

func (c *Checker) Check(pass *analysis.Pass) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}
	insp.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		c.checkCallExpr(call, pass)
	})
}

func (c *Checker) checkCallExpr(call *ast.CallExpr, pass *analysis.Pass) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	if !c.detector.IsLogMethod(sel.Sel.Name) {
		return
	}

	if !c.detector.IsLogger(pass, sel.X) {
		return
	}

	c.executeRules(call, pass)
}

func (c *Checker) executeRules(call *ast.CallExpr, pass *analysis.Pass) {
	if len(call.Args) == 0 {
		return
	}

	msgExpr := call.Args[0]
	msgValue := extractStringValue(msgExpr)

	ctx := &rules.CheckContext{
		MsgExpr: msgExpr,
		Msg:     msgValue,
	}

	for _, rule := range c.rules {
		result := rule.Check(ctx)
		if !result.Passed {
			c.reportViolation(pass, msgExpr, result)
		}
	}
}

func (c *Checker) reportViolation(pass *analysis.Pass, expr ast.Expr, result *rules.RuleResult) {
	diag := analysis.Diagnostic{
		Pos:     expr.Pos(),
		End:     expr.End(),
		Message: result.Message,
	}

	if result.SuggestedFix != nil {
		newText := result.SuggestedFix.NewText
		if isStringLiteral(expr) {
			newText = `"` + newText + `"`
		}
		diag.SuggestedFixes = []analysis.SuggestedFix{
			{
				Message: result.SuggestedFix.Message,
				TextEdits: []analysis.TextEdit{
					{
						Pos:     expr.Pos(),
						End:     expr.End(),
						NewText: []byte(newText),
					},
				},
			},
		}
	}

	pass.Report(diag)
}

func extractStringValue(expr ast.Expr) string {
	if expr == nil {
		return ""
	}

	switch v := expr.(type) {
	case *ast.BasicLit:
		if v.Kind.String() == "STRING" && len(v.Value) >= 2 {
			return v.Value[1 : len(v.Value)-1]
		}
	case *ast.BinaryExpr:
		return extractConcatenatedValue(v)
	}
	return ""
}

func extractConcatenatedValue(expr *ast.BinaryExpr) string {
	var parts []string
	var walk func(ast.Expr)

	walk = func(e ast.Expr) {
		switch v := e.(type) {
		case *ast.BasicLit:
			if v.Kind.String() == "STRING" && len(v.Value) >= 2 {
				parts = append(parts, v.Value[1:len(v.Value)-1])
			}
		case *ast.BinaryExpr:
			walk(v.X)
			walk(v.Y)
		}
	}
	walk(expr)
	return strings.Join(parts, "")
}

func isStringLiteral(expr ast.Expr) bool {
	_, ok := expr.(*ast.BasicLit)
	return ok
}
