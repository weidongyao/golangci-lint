package goanalysis

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/analysis"

	"github.com/golangci/golangci-lint/pkg/lint/linter"
	"github.com/golangci/golangci-lint/pkg/result"
)

type MetaLinter struct {
	linters []*Linter
}

func NewMetaLinter(linters []*Linter) *MetaLinter {
	return &MetaLinter{linters: linters}
}

func (ml MetaLinter) Name() string {
	return "goanalysis_metalinter"
}

func (ml MetaLinter) Desc() string {
	return ""
}

func (ml MetaLinter) Run(ctx context.Context, lintCtx *linter.Context) ([]result.Issue, error) {
	for _, linter := range ml.linters {
		if err := analysis.Validate(linter.analyzers); err != nil {
			return nil, errors.Wrapf(err, "failed to validate analyzers of %s", linter.Name())
		}
	}

	for _, linter := range ml.linters {
		if err := linter.configure(); err != nil {
			return nil, errors.Wrapf(err, "failed to configure analyzers of %s", linter.Name())
		}
	}

	var allAnalyzers []*analysis.Analyzer
	analyzerToLinterName := map[*analysis.Analyzer]string{}
	for _, linter := range ml.linters {
		if linter.contextSetter != nil {
			linter.contextSetter(lintCtx)
		}

		allAnalyzers = append(allAnalyzers, linter.analyzers...)
		for _, a := range linter.analyzers {
			analyzerToLinterName[a] = linter.Name()
		}
	}

	loadMode := LoadModeNone
	for _, linter := range ml.linters {
		if linter.loadMode > loadMode {
			loadMode = linter.loadMode
		}
	}

	isTypecheckMode := false
	for _, linter := range ml.linters {
		if linter.isTypecheckMode {
			isTypecheckMode = true
			break
		}
	}

	useOriginalPackages := false
	for _, linter := range ml.linters {
		if linter.useOriginalPackages {
			useOriginalPackages = true
			break
		}
	}

	runner := newRunner("metalinter", lintCtx.Log.Child("goanalysis"), lintCtx.PkgCache, lintCtx.LoadGuard, loadMode)

	pkgs := lintCtx.Packages
	if useOriginalPackages {
		pkgs = lintCtx.OriginalPackages
	}

	diags, errs := runner.run(allAnalyzers, pkgs)

	buildAllIssues := func() []result.Issue {
		linterNameBuilder := func(diag *Diagnostic) string { return analyzerToLinterName[diag.Analyzer] }
		issues := buildIssues(diags, linterNameBuilder)

		for _, linter := range ml.linters {
			if linter.issuesReporter != nil {
				issues = append(issues, linter.issuesReporter(lintCtx)...)
			}
		}
		return issues
	}

	if isTypecheckMode {
		issues, err := buildIssuesFromErrorsForTypecheckMode(errs, lintCtx)
		if err != nil {
			return nil, err
		}
		return append(issues, buildAllIssues()...), nil
	}

	// Don't print all errs: they can duplicate.
	if len(errs) != 0 {
		return nil, errs[0]
	}

	return buildAllIssues(), nil
}
