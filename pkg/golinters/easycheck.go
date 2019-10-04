/*
@Time : 2019-09-29 20:58
@Author : yaoweidong
@File : ezcheck
@Comment:
*/
package golinters

import (
	"bytes"
	"github.com/golangci/golangci-lint/pkg/golinters/goanalysis"
	"go/ast"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"strings"
)

func NewEasyCheck() *goanalysis.Linter {
	return goanalysis.NewLinter(
		"easycheck",
		"easycheck examines Go source code and reports suspicious constructs.",
		generateEasyCheckAnalyzer(),
		nil,
	)
}

func generateEasyCheckAnalyzer() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		getGoroutineCatchPanicAnalyzer(),
		getGormBeginAnalyzer(),
	}
}

//检查goroutine是否catch panic
func getGoroutineCatchPanicAnalyzer() *analysis.Analyzer {
	runner := func(pass *analysis.Pass) (interface{}, error) {
		// get the inspector. This will not panic because inspect.CatchPanicAnalyzer is part
		// of `Requires`. go/analysis will populate the `pass.ResultOf` map with
		// the prerequisite analyzers.
		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		// the inspector has a `filter` feature that enables type-based filtering
		// The anonymous function will be only called for the ast nodes whose type
		// matches an element in the filter
		nodeFilter := []ast.Node{
			(*ast.GoStmt)(nil),
		}
		inspect.Preorder(nodeFilter, func(node ast.Node) {
			goStmt := node.(*ast.GoStmt)
			funclit, ok := goStmt.Call.Fun.(*ast.FuncLit)
			if !ok {
				return
			}
			catchPanic := false
			for _, stmt := range funclit.Body.List {
				if deferStmt, ok := stmt.(*ast.DeferStmt); ok {
					if selector, ok := deferStmt.Call.Fun.(*ast.SelectorExpr); ok {
						if selector.Sel.Name == "CatchPanic" {
							catchPanic = true
						}
					}
				}
			}
			if !catchPanic {
				pass.Reportf(goStmt.Pos(), "goroutine miss catch panic %q",
					render(pass.Fset, goStmt))
			}
		})
		return nil, nil
	}

	return &analysis.Analyzer{
		Name:     "catcherlint",
		Doc:      "examine if catch panic in goroutine",
		Run:      runner,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

//检查gorm begin后是否判断err
func getGormBeginAnalyzer() *analysis.Analyzer {
	runner := func(pass *analysis.Pass) (interface{}, error) {
		// get the inspector. This will not panic because inspect.CatchPanicAnalyzer is part
		// of `Requires`. go/analysis will populate the `pass.ResultOf` map with
		// the prerequisite analyzers.
		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		// the inspector has a `filter` feature that enables type-based filtering
		// The anonymous function will be only called for the ast nodes whose type
		// matches an element in the filter
		nodeFilter := []ast.Node{
			(*ast.BlockStmt)(nil),
		}
		inspect.Preorder(nodeFilter, func(node ast.Node) {
			blockStmt := node.(*ast.BlockStmt)
			hasBeginStmt := false  //是否是begin赋值语句
			hasExamineErr := false //判断后边3行内是否有判断err为nil
			var beginStmt *ast.AssignStmt
			var beginPos token.Pos
			for index, stmt := range blockStmt.List {
				if assignStmt, ok := stmt.(*ast.AssignStmt); ok {
					paramName := ""
					lhs := assignStmt.Lhs[0]
					if ident, ok := lhs.(*ast.Ident); ok {
						paramName = ident.Name
					}
					rhs := assignStmt.Rhs[0]
					if callExpr, ok := rhs.(*ast.CallExpr); ok {
						if selector, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
							if strings.HasSuffix(pass.TypesInfo.TypeOf(selector.X).String(), "gorm.DB") {
								if selector.Sel.Name == "Begin" {
									hasBeginStmt = true
									beginStmt = assignStmt
									beginPos = assignStmt.Pos()
								}
							}
						}
					}
					if hasBeginStmt {
						for i := index + 1; i < len(blockStmt.List) && i-index <= 3; i++ {
							nextStmt := blockStmt.List[i]
							if ifStmt, ok := nextStmt.(*ast.IfStmt); !ok {
								continue
							} else {
								if binaryExpr, ok := ifStmt.Cond.(*ast.BinaryExpr); !ok {
									continue
								} else {
									if selector, ok := binaryExpr.X.(*ast.SelectorExpr); !ok {
										continue
									} else {
										if ident, ok := selector.X.(*ast.Ident); ok {
											if ident.Name == paramName && selector.Sel.Name == "Error" && binaryExpr.Op.String() == "!=" {
												hasExamineErr = true
												break
											}
										}
									}
								}
							}
						}
					}
				}
			}
			if hasBeginStmt && !hasExamineErr {
				pass.Reportf(beginPos, "gorm Begin without handle err %q",
					render(pass.Fset, beginStmt))
			}
		})
		return nil, nil
	}
	return &analysis.Analyzer{
		Name:     "gormlint",
		Doc:      "examine if check err after begin invoking",
		Run:      runner,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

// render returns the pretty-print of the given node
func render(fset *token.FileSet, x interface{}) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, x); err != nil {
		panic(err)
	}
	return buf.String()
}
