package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/shinagawa-web/gomarklint/internal/config"
	"github.com/shinagawa-web/gomarklint/internal/parser"
	"github.com/shinagawa-web/gomarklint/internal/rule"
	"github.com/spf13/cobra"
)

var minHeadingLevel int
var checkLinks bool
var skipLinkPatterns []string
var configFilePath string
var outputFormat string

var rootCmd = &cobra.Command{
	Use:   "gomarklint [files or directories]",
	Short: "A fast markdown linter written in Go",
	Long:  "gomarklint checks markdown files for common issues like heading structure, blank lines, and more.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println()

		start := time.Now()
		cfg := config.Default()
		if _, err := os.Stat(configFilePath); err == nil {
			c, err := config.LoadConfig(configFilePath)
			if err != nil {
				return err
			}
			cfg = c
		}

		if cmd.Flags().Changed("min-heading") {
			cfg.MinHeadingLevel = minHeadingLevel
		}
		if cmd.Flags().Changed("check-links") {
			cfg.CheckLinks = checkLinks
		}
		if cmd.Flags().Changed("skip-link-patterns") {
			cfg.SkipLinkPatterns = skipLinkPatterns
		}

		if cmd.Flags().Changed("output") {
			cfg.OutputFormat = outputFormat
		}
		outputFormat = cfg.OutputFormat

		if outputFormat != "text" && outputFormat != "json" {
			return fmt.Errorf("invalid --output value: %q (must be 'text' or 'json')", outputFormat)
		}

		minHeadingLevel = cfg.MinHeadingLevel
		checkLinks = cfg.CheckLinks
		skipLinkPatterns = cfg.SkipLinkPatterns

		if len(args) == 0 {
			return fmt.Errorf("please provide a markdown file or directory")
		}

		files, err := parser.ExpandPaths(args, cfg.Ignore)
		if err != nil {
			return fmt.Errorf("failed to expand paths: %w", err)
		}
		compiledPatterns := []*regexp.Regexp{}
		if checkLinks {
			for _, pat := range skipLinkPatterns {
				re, err := regexp.Compile(pat)
				if err != nil {
					log.Printf("Invalid skip-link-pattern: %s (error: %v)", pat, err)
					continue
				}
				compiledPatterns = append(compiledPatterns, re)
			}
		}

		totalErrors := 0
		totalLines := 0
		results := map[string][]rule.LintError{}
		orderedPaths := make([]string, 0, len(files))

		for _, path := range files {
			content, err := parser.ReadFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", path, err)
				continue
			}
			errors, lineCount := collectErrors(path, content, cfg, compiledPatterns)
			results[path] = errors
			orderedPaths = append(orderedPaths, path)
			totalErrors += len(errors)
			totalLines += lineCount
		}

		elapsed := time.Since(start)

		if outputFormat == "json" {
			printJSONOutput(results, len(files), totalLines, totalErrors, elapsed)
		} else {
			printHumanOutput(orderedPaths, results, len(files), totalLines, totalErrors, elapsed)
		}
		return nil
	},
}

func collectErrors(path string, content string, cfg config.Config, patterns []*regexp.Regexp) ([]rule.LintError, int) {
	var allErrors []rule.LintError
	allErrors = append(allErrors, rule.CheckHeadingLevels(path, content, cfg.MinHeadingLevel)...)
	allErrors = append(allErrors, rule.CheckFinalBlankLine(path, content)...)
	allErrors = append(allErrors, rule.CheckUnclosedCodeBlocks(path, content)...)
	allErrors = append(allErrors, rule.CheckDuplicateHeadings(path, content)...)
	allErrors = append(allErrors, rule.CheckEmptyAltText(path, content)...)
	if cfg.CheckLinks {
		allErrors = append(allErrors, rule.CheckExternalLinks(path, content, patterns)...)
	}

	sort.Slice(allErrors, func(i, j int) bool {
		return allErrors[i].Line < allErrors[j].Line
	})

	lineCount := strings.Count(content, "\n") + 1
	return allErrors, lineCount
}

func printJSONOutput(results map[string][]rule.LintError, totalFiles, totalLines, totalErrors int, duration time.Duration) {
	output := struct {
		Files       int                         `json:"files"`
		Lines       int                         `json:"lines"`
		Errors      int                         `json:"errors"`
		ElapsedMS   int64                       `json:"elapsed_ms"`
		ErrorDetail map[string][]rule.LintError `json:"details"`
	}{
		Files:       totalFiles,
		Lines:       totalLines,
		Errors:      totalErrors,
		ElapsedMS:   duration.Milliseconds(),
		ErrorDetail: results,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(output); err != nil {
		log.Fatalf("failed to write JSON output: %v", err)
	}
}

func printHumanOutput(
	orderedPaths []string,
	results map[string][]rule.LintError,
	totalFiles, totalLines, totalErrors int,
	duration time.Duration,
) {
	red := "\033[31m"
	green := "\033[32m"
	gray := "\033[90m"
	reset := "\033[0m"

	for _, path := range orderedPaths {
		errors := results[path]
		if len(errors) == 0 {
			continue
		}
		fmt.Printf("Errors in %s:\n", path)
		for _, e := range errors {
			fmt.Printf("  %s:%d: %s\n", e.File, e.Line, e.Message)
		}
		fmt.Println()
	}

	if totalErrors > 0 {
		fmt.Printf("\n%s✖ %d issues found%s\n", red, totalErrors, reset)
	} else {
		fmt.Printf("\n%s✔ No issues found%s\n", green, reset)
	}

	if duration < time.Second {
		fmt.Printf("%s✓%s Checked %d file(s), %d line(s) in %s%dms%s\n",
			green, reset, totalFiles, totalLines, gray, duration.Milliseconds(), reset)
	} else {
		fmt.Printf("%s✓%s Checked %d file(s), %d line(s) in %s%.1fs%s\n",
			green, reset, totalFiles, totalLines, gray, duration.Seconds(), reset)
	}
}

func init() {
	rootCmd.Flags().StringVar(&configFilePath, "config", ".gomarklint.json", "path to config file (default: .gomarklint.json)")

	rootCmd.Flags().IntVar(&minHeadingLevel, "min-heading", 2, "minimum heading level to start from (default: 2)")
	rootCmd.Flags().BoolVar(&checkLinks, "check-links", false, "enable external link checking")
	rootCmd.Flags().StringArrayVar(&skipLinkPatterns, "skip-link-patterns", nil, "patterns of URLs to skip link checking")
	rootCmd.Flags().StringVar(&outputFormat, "output", "text", "output format: text or json")

	rootCmd.AddCommand(initCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
