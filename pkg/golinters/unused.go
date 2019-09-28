package golinters

import (
	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/unused"

	"github.com/golangci/golangci-lint/pkg/lint/linter"
	"github.com/golangci/golangci-lint/pkg/result"

	"github.com/golangci/golangci-lint/pkg/golinters/goanalysis"
)

func NewUnused() *goanalysis.Linter {
	u := unused.NewChecker(false)
	analyzers := []*analysis.Analyzer{u.Analyzer()}
	setAnalyzersGoVersion(analyzers)

	const name = "unused"
	linter := goanalysis.NewLinter(
		name,
		"Checks Go code for unused constants, variables, functions and types",
		analyzers,
		nil,
	).WithIssuesReporter(func(lintCtx *linter.Context) []result.Issue {
		var issues []result.Issue
		for _, ur := range u.Result() {
			p := u.ProblemObject(lintCtx.Packages[0].Fset, ur)
			issues = append(issues, result.Issue{
				FromLinter: name,
				Text:       p.Message,
				Pos:        p.Pos,
			})
		}
		return issues
	}).WithContextSetter(func(lintCtx *linter.Context) {
		u.WholeProgram = lintCtx.Settings().Unused.CheckExported
	}).WithLoadMode(goanalysis.LoadModeWholeProgram)
	linter.UseOriginalPackages()
	return linter
}
