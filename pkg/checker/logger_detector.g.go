package checker

import (
	"go/ast"
	"go/types"
	"slices"

	"golang.org/x/tools/go/analysis"
)

const (
	slogPackage = "log/slog"
	zapPackage  = "go.uber.org/zap"
)

var logMethods = map[string][]string{
	slogPackage: {"Debug", "Info", "Warn", "Error", "Log", "LogAttrs"},
	zapPackage:  {"Debug", "Info", "Warn", "Error", "DPanic", "Panic", "Fatal"},
}

type LoggerDetector struct {
	allMethods []string
}

func NewLoggerDetector() *LoggerDetector {
	detector := &LoggerDetector{}
	for _, methods := range logMethods {
		detector.allMethods = append(detector.allMethods, methods...)
	}
	return detector
}

func (d *LoggerDetector) IsLogMethod(methodName string) bool {
	return slices.Contains(d.allMethods, methodName)
}

func (d *LoggerDetector) IsLogger(pass *analysis.Pass, expr ast.Expr) bool {
	ident := selToIdent(expr)
	if ident == nil {
		return false
	}

	obj := pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return false
	}

	typ := obj.Type()
	if typ == nil {
		return false
	}

	if ptr, ok := typ.(*types.Pointer); ok {
		if named, ok := ptr.Elem().(*types.Named); ok {
			pkg := named.Obj().Pkg()
			if pkg != nil && pkg.Path() == zapPackage {
				return named.Obj().Name() == "Logger"
			}
		}
	}

	if pkgName, ok := obj.(*types.PkgName); ok {
		return pkgName.Imported().Path() == slogPackage
	}

	return false
}

func selToIdent(expr ast.Expr) *ast.Ident {
	switch v := expr.(type) {
	case *ast.Ident:
		return v
	case *ast.SelectorExpr:
		if ident, ok := v.X.(*ast.Ident); ok {
			return ident
		}
	}
	return nil
}
