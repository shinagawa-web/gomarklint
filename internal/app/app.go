package app

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/shinagawa-web/gomarklint/v2/internal/config"
	"github.com/shinagawa-web/gomarklint/v2/internal/file"
	"github.com/shinagawa-web/gomarklint/v2/internal/linter"
	"github.com/shinagawa-web/gomarklint/v2/internal/output"
	"github.com/shinagawa-web/gomarklint/v2/internal/rule"
)

// ErrLintViolations is returned when error-severity violations are found.
var ErrLintViolations = errors.New("lint violations found")

// Options holds the parameters for a lint run.
type Options struct {
	ConfigPath   string              // path to config file
	Args         []string            // files/dirs to lint
	OutputFormat string              // overrides config if non-empty
	MinSeverity  config.RuleSeverity // overrides config if non-empty
}

// Run loads config, lints files, and writes results to w.
// Returns ErrLintViolations if any error-severity violations are found.
func Run(w io.Writer, opts Options) error {
	start := time.Now()

	cfg, err := config.LoadOrDefault(opts.ConfigPath)
	if err != nil {
		return err
	}

	if opts.OutputFormat != "" {
		cfg.OutputFormat = opts.OutputFormat
	}
	if opts.MinSeverity != "" {
		cfg.MinSeverity = opts.MinSeverity
	}

	if err := config.Validate(cfg); err != nil {
		return err
	}

	args := opts.Args
	if len(args) == 0 {
		if len(cfg.Include) > 0 {
			args = cfg.Include
		} else {
			return fmt.Errorf("please provide a markdown file or directory (or set 'include' in .gomarklint.json)")
		}
	}

	files := file.ExpandPaths(args, cfg.Ignore)

	lint := linter.New(cfg)
	result := lint.Run(files)

	if err := formatOutput(w, cfg, result, len(files)-len(result.FailedFiles), time.Since(start)); err != nil {
		return err
	}

	if result.TotalErrors > 0 {
		return ErrLintViolations
	}
	return nil
}

func formatOutput(w io.Writer, cfg config.Config, result *linter.Result, fileCount int, duration time.Duration) error {
	var formatter output.Formatter
	if cfg.OutputFormat == "json" {
		formatter = output.NewJSONFormatter()
	} else {
		formatter = output.NewTextFormatter()
	}

	linksChecked := &result.TotalLinksChecked
	if !cfg.IsEnabled("external-link") {
		linksChecked = nil
	}

	details, errCount, warnCount := filterBySeverity(result.Errors, cfg.MinSeverity)

	outputResult := &output.Result{
		Files:        fileCount,
		Lines:        result.TotalLines,
		Total:        errCount + warnCount,
		Warnings:     warnCount,
		LinksChecked: linksChecked,
		Duration:     duration,
		Details:      details,
		OrderedPaths: result.OrderedPaths,
	}

	return formatter.Format(w, outputResult)
}

func filterBySeverity(details map[string][]rule.LintError, minSev config.RuleSeverity) (map[string][]rule.LintError, int, int) {
	filtered := make(map[string][]rule.LintError, len(details))
	errCount := 0
	warnCount := 0
	for path, errs := range details {
		var kept []rule.LintError
		for _, e := range errs {
			if minSev == config.SeverityError && e.Severity == string(config.SeverityWarning) {
				continue
			}
			kept = append(kept, e)
			if e.Severity == string(config.SeverityWarning) {
				warnCount++
			} else {
				errCount++
			}
		}
		if len(kept) > 0 {
			filtered[path] = kept
		}
	}
	return filtered, errCount, warnCount
}
