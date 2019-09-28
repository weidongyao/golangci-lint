package lint

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/golangci/golangci-lint/internal/errorutil"
	"github.com/golangci/golangci-lint/pkg/lint/lintersdb"

	"github.com/golangci/golangci-lint/pkg/fsutils"

	"github.com/golangci/golangci-lint/pkg/config"
	"github.com/golangci/golangci-lint/pkg/goutil"
	"github.com/golangci/golangci-lint/pkg/lint/astcache"
	"github.com/golangci/golangci-lint/pkg/lint/linter"
	"github.com/golangci/golangci-lint/pkg/logutils"
	"github.com/golangci/golangci-lint/pkg/packages"
	"github.com/golangci/golangci-lint/pkg/result"
	"github.com/golangci/golangci-lint/pkg/result/processors"
	"github.com/golangci/golangci-lint/pkg/timeutils"
)

type Runner struct {
	Processors []processors.Processor
	Log        logutils.Log
}

func NewRunner(astCache *astcache.Cache, cfg *config.Config, log logutils.Log, goenv *goutil.Env,
	lineCache *fsutils.LineCache, dbManager *lintersdb.Manager) (*Runner, error) {
	icfg := cfg.Issues
	excludePatterns := icfg.ExcludePatterns
	if icfg.UseDefaultExcludes {
		excludePatterns = append(excludePatterns, config.GetDefaultExcludePatternsStrings()...)
	}

	var excludeTotalPattern string
	if len(excludePatterns) != 0 {
		excludeTotalPattern = fmt.Sprintf("(%s)", strings.Join(excludePatterns, "|"))
	}

	skipFilesProcessor, err := processors.NewSkipFiles(cfg.Run.SkipFiles)
	if err != nil {
		return nil, err
	}

	skipDirs := cfg.Run.SkipDirs
	if cfg.Run.UseDefaultSkipDirs {
		skipDirs = append(skipDirs, packages.StdExcludeDirRegexps...)
	}
	skipDirsProcessor, err := processors.NewSkipDirs(skipDirs, log.Child("skip dirs"), cfg.Run.Args)
	if err != nil {
		return nil, err
	}

	var excludeRules []processors.ExcludeRule
	for _, r := range icfg.ExcludeRules {
		excludeRules = append(excludeRules, processors.ExcludeRule{
			Text:    r.Text,
			Source:  r.Source,
			Path:    r.Path,
			Linters: r.Linters,
		})
	}

	return &Runner{
		Processors: []processors.Processor{
			processors.NewCgo(goenv),
			processors.NewFilenameUnadjuster(astCache, log.Child("filename_unadjuster")), // must go after Cgo
			processors.NewPathPrettifier(), // must be before diff, nolint and exclude autogenerated processor at least
			skipFilesProcessor,
			skipDirsProcessor, // must be after path prettifier

			processors.NewAutogeneratedExclude(astCache),
			processors.NewIdentifierMarker(), // must be before exclude because users see already marked output and configure excluding by it
			processors.NewExclude(excludeTotalPattern),
			processors.NewExcludeRules(excludeRules, lineCache, log.Child("exclude_rules")),
			processors.NewNolint(astCache, log.Child("nolint"), dbManager),

			processors.NewUniqByLine(cfg),
			processors.NewDiff(icfg.Diff, icfg.DiffFromRevision, icfg.DiffPatchFilePath),
			processors.NewMaxPerFileFromLinter(cfg),
			processors.NewMaxSameIssues(icfg.MaxSameIssues, log.Child("max_same_issues"), cfg),
			processors.NewMaxFromLinter(icfg.MaxIssuesPerLinter, log.Child("max_from_linter"), cfg),
			processors.NewSourceCode(lineCache, log.Child("source_code")),
			processors.NewPathShortener(),
		},
		Log: log,
	}, nil
}

type lintRes struct {
	linter *linter.Config
	err    error
	issues []result.Issue
}

func (r *Runner) runLinterSafe(ctx context.Context, lintCtx *linter.Context,
	lc *linter.Config) (ret []result.Issue, err error) {
	defer func() {
		if panicData := recover(); panicData != nil {
			if pe, ok := panicData.(*errorutil.PanicError); ok {
				// Don't print stacktrace from goroutines twice
				lintCtx.Log.Warnf("Panic: %s: %s", pe, pe.Stack())
			} else {
				err = fmt.Errorf("panic occurred: %s", panicData)
				r.Log.Warnf("Panic stack trace: %s", debug.Stack())
			}
		}
	}()

	specificLintCtx := *lintCtx
	specificLintCtx.Log = r.Log.Child(lc.Name())
	issues, err := lc.Linter.Run(ctx, &specificLintCtx)
	if err != nil {
		return nil, err
	}

	for _, i := range issues {
		i.FromLinter = lc.Name()
	}

	return issues, nil
}

func (r Runner) runWorker(ctx context.Context, lintCtx *linter.Context,
	tasksCh <-chan *linter.Config, lintResultsCh chan<- lintRes, name string) {

	for {
		select {
		case <-ctx.Done():
			return
		case lc, ok := <-tasksCh:
			if !ok {
				return
			}
			if ctx.Err() != nil {
				// XXX: if check it in only int a select
				// it's possible to not enter to this case until tasksCh is empty.
				return
			}
			var issues []result.Issue
			var err error

			lintResultsCh <- lintRes{
				linter: lc,
				err:    err,
				issues: issues,
			}
		}
	}
}

func (r Runner) logWorkersStat(workersFinishTimes []time.Time) {
	lastFinishTime := workersFinishTimes[0]
	for _, t := range workersFinishTimes {
		if t.After(lastFinishTime) {
			lastFinishTime = t
		}
	}

	logStrings := []string{}
	for i, t := range workersFinishTimes {
		if t.Equal(lastFinishTime) {
			continue
		}

		logStrings = append(logStrings, fmt.Sprintf("#%d: %s", i+1, lastFinishTime.Sub(t)))
	}

	r.Log.Infof("Workers idle times: %s", strings.Join(logStrings, ", "))
}

type processorStat struct {
	inCount  int
	outCount int
}

func (r Runner) processLintResults(inIssues []result.Issue) []result.Issue {
	sw := timeutils.NewStopwatch("processing", r.Log)

	var issuesBefore, issuesAfter int
	statPerProcessor := map[string]processorStat{}

	var outIssues []result.Issue
	if len(inIssues) != 0 {
		issuesBefore += len(inIssues)
		outIssues = r.processIssues(inIssues, sw, statPerProcessor)
		issuesAfter += len(outIssues)
	}

	// finalize processors: logging, clearing, no heavy work here

	for _, p := range r.Processors {
		p := p
		sw.TrackStage(p.Name(), func() {
			p.Finish()
		})
	}

	if issuesBefore != issuesAfter {
		r.Log.Infof("Issues before processing: %d, after processing: %d", issuesBefore, issuesAfter)
	}
	r.printPerProcessorStat(statPerProcessor)
	sw.PrintStages()

	return outIssues
}

func (r Runner) printPerProcessorStat(stat map[string]processorStat) {
	parts := make([]string, 0, len(stat))
	for name, ps := range stat {
		if ps.inCount != 0 {
			parts = append(parts, fmt.Sprintf("%s: %d/%d", name, ps.outCount, ps.inCount))
		}
	}
	if len(parts) != 0 {
		r.Log.Infof("Processors filtering stat (out/in): %s", strings.Join(parts, ", "))
	}
}

func (r Runner) Run(ctx context.Context, linters []*linter.Config, lintCtx *linter.Context) []result.Issue {
	sw := timeutils.NewStopwatch("linters", r.Log)
	defer sw.Print()

	var issues []result.Issue
	for _, lc := range linters {
		sw.TrackStage(lc.Name(), func() {
			linterIssues, err := r.runLinterSafe(ctx, lintCtx, lc)
			if err != nil {
				r.Log.Warnf("Can't run linter %s: %s", lc.Linter.Name(), err)
				return
			}
			issues = append(issues, linterIssues...)
		})
	}

	return r.processLintResults(issues)
}

func (r *Runner) processIssues(issues []result.Issue, sw *timeutils.Stopwatch, statPerProcessor map[string]processorStat) []result.Issue {
	for _, p := range r.Processors {
		var newIssues []result.Issue
		var err error
		p := p
		sw.TrackStage(p.Name(), func() {
			newIssues, err = p.Process(issues)
		})

		if err != nil {
			r.Log.Warnf("Can't process result by %s processor: %s", p.Name(), err)
		} else {
			stat := statPerProcessor[p.Name()]
			stat.inCount += len(issues)
			stat.outCount += len(newIssues)
			statPerProcessor[p.Name()] = stat
			issues = newIssues
		}

		if issues == nil {
			issues = []result.Issue{}
		}
	}

	return issues
}
