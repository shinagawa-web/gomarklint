package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/shinagawa-web/gomarklint/v2/internal/config"
	"github.com/shinagawa-web/gomarklint/v2/internal/file"
	"github.com/shinagawa-web/gomarklint/v2/internal/linter"
	"github.com/shinagawa-web/gomarklint/v2/internal/output"
	"github.com/shinagawa-web/gomarklint/v2/internal/rule"
)

// ErrLintViolations is returned when lint violations are found.
var ErrLintViolations = errors.New("lint violations found")

var configFilePath string
var outputFormat string
var minSeverity string

var rootCmd = &cobra.Command{
	Use:   "gomarklint [files or directories]",
	Short: "A fast markdown linter written in Go",
	Long:  "gomarklint checks markdown files for common issues like heading structure, blank lines, and more.",
	Args:  cobra.MinimumNArgs(0),
	RunE:  runLint,
}

func runLint(cmd *cobra.Command, args []string) error {
	fmt.Println()
	start := time.Now()

	// Load and merge configuration
	cfg, err := config.LoadOrDefault(configFilePath)
	if err != nil {
		return err
	}

	flags := config.FlagValues{
		OutputFormat: outputFormat,
		MinSeverity:  minSeverity,
	}
	cfg = config.MergeFlags(cfg, cmd, flags)

	if err := config.Validate(cfg); err != nil {
		return err
	}

	// Determine files to check
	if len(args) == 0 {
		if len(cfg.Include) > 0 {
			args = cfg.Include
		} else {
			return fmt.Errorf("please provide a markdown file or directory (or set 'include' in .gomarklint.json)")
		}
	}

	files, err := file.ExpandPaths(args, cfg.Ignore)
	if err != nil {
		return fmt.Errorf("failed to expand paths: %w", err)
	}

	// Run linter
	lint, err := linter.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create linter: %w", err)
	}

	result, err := lint.Run(files)
	if err != nil {
		return err
	}

	// Format and output results
	if err := formatOutput(cfg, result, len(files), time.Since(start)); err != nil {
		return err
	}

	// Exit 1 only when there are severity=error violations (warnings never fail)
	if result.TotalErrors > 0 {
		return ErrLintViolations
	}
	return nil
}

func formatOutput(cfg config.Config, result *linter.Result, fileCount int, duration time.Duration) error {
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

	// Filter violations by MinSeverity: "error" hides warnings from output
	details, errCount, warnCount := filterBySeverity(result.Errors, cfg.MinSeverity)

	outputResult := &output.Result{
		Files:        fileCount,
		Lines:        result.TotalLines,
		Errors:       errCount + warnCount,
		Warnings:     warnCount,
		LinksChecked: linksChecked,
		Duration:     duration,
		Details:      details,
		OrderedPaths: result.OrderedPaths,
	}

	return formatter.Format(os.Stdout, outputResult)
}

// filterBySeverity filters violations by minimum severity and returns filtered details,
// error count, and warning count.
func filterBySeverity(details map[string][]rule.LintError, minSev config.RuleSeverity) (map[string][]rule.LintError, int, int) {
	filtered := make(map[string][]rule.LintError, len(details))
	errCount := 0
	warnCount := 0
	for path, errs := range details {
		var kept []rule.LintError
		for _, e := range errs {
			if minSev == config.SeverityError && e.Severity == "warning" {
				continue
			}
			kept = append(kept, e)
			if e.Severity == "warning" {
				warnCount++
			} else {
				errCount++
			}
		}
		filtered[path] = kept
	}
	return filtered, errCount, warnCount
}

func init() {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	rootCmd.Flags().StringVar(&configFilePath, "config", ".gomarklint.json", "path to config file (default: .gomarklint.json)")
	rootCmd.Flags().StringVar(&outputFormat, "output", "text", "output format: text or json")
	rootCmd.Flags().StringVar(&minSeverity, "severity", "warning", "minimum severity to report: warning or error")

	rootCmd.AddCommand(initCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
