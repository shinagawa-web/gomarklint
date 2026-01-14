package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/shinagawa-web/gomarklint/internal/config"
	"github.com/shinagawa-web/gomarklint/internal/parser"
	"github.com/shinagawa-web/gomarklint/internal/rule"
)

var minHeadingLevel int
var enableLinkCheck bool
var skipLinkPatterns []string
var configFilePath string
var outputFormat string
var enableHeadingLevelCheck bool
var enableDuplicateHeadingCheck bool
var enableNoMultipleBlankLinesCheck bool

var rootCmd = &cobra.Command{
	Use:   "gomarklint [files or directories]",
	Short: "A fast markdown linter written in Go",
	Long:  "gomarklint checks markdown files for common issues like heading structure, blank lines, and more.",
	Args:  cobra.MinimumNArgs(0),
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
		if cmd.Flags().Changed("enable-link-check") {
			cfg.EnableLinkCheck = enableLinkCheck
		}
		if cmd.Flags().Changed("enable-heading-level-check") {
			cfg.EnableHeadingLevelCheck = enableHeadingLevelCheck
		}
		if cmd.Flags().Changed("enable-duplicate-heading-check") {
			cfg.EnableDuplicateHeadingCheck = enableDuplicateHeadingCheck
		}
		if cmd.Flags().Changed("enable-no-multiple-blank-lines-check") {
			cfg.EnableNoMultipleBlankLinesCheck = enableNoMultipleBlankLinesCheck
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
		enableLinkCheck = cfg.EnableLinkCheck
		skipLinkPatterns = cfg.SkipLinkPatterns
		enableHeadingLevelCheck = cfg.EnableHeadingLevelCheck
		enableDuplicateHeadingCheck = cfg.EnableDuplicateHeadingCheck
		enableNoMultipleBlankLinesCheck = cfg.EnableNoMultipleBlankLinesCheck

		if len(args) == 0 {
			if len(cfg.Include) > 0 {
				args = cfg.Include
			} else {
				return fmt.Errorf("please provide a markdown file or directory (or set 'include' in .gomarklint.json)")
			}
		}

		files, err := parser.ExpandPaths(args, cfg.Ignore)
		if err != nil {
			return fmt.Errorf("failed to expand paths: %w", err)
		}
		compiledPatterns := []*regexp.Regexp{}
		if enableLinkCheck {
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
		totalLinksChecked := 0
		results := map[string][]rule.LintError{}
		orderedPaths := make([]string, 0, len(files))

		var mu sync.Mutex
		var wg sync.WaitGroup

		urlCache := &sync.Map{}

		for _, path := range files {
			wg.Add(1)
			go func(p string) {
				defer wg.Done()

				content, err := parser.ReadFile(p)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", p, err)
					return
				}
				errors, lineCount, linksChecked := collectErrors(p, content, cfg, compiledPatterns, urlCache)

				mu.Lock()
				results[p] = errors
				orderedPaths = append(orderedPaths, p)
				totalErrors += len(errors)
				totalLines += lineCount
				totalLinksChecked += linksChecked
				mu.Unlock()
			}(path)
		}

		wg.Wait()

		// Sort paths to ensure consistent output order
		sort.Strings(orderedPaths)

		elapsed := time.Since(start)

		if outputFormat == "json" {
			printJSONOutput(results, len(files), totalLines, totalErrors, totalLinksChecked, cfg.EnableLinkCheck, elapsed)
		} else {
			printHumanOutput(orderedPaths, results, len(files), totalLines, totalErrors, totalLinksChecked, cfg.EnableLinkCheck, elapsed)
		}

		onCI := os.Getenv("GITHUB_ACTIONS") == "true"
		if totalErrors > 0 {
			if onCI {
				return errors.New("")
			}
			return nil
		}
		return nil
	},
}

func collectErrors(path string, content string, cfg config.Config, patterns []*regexp.Regexp, urlCache *sync.Map) ([]rule.LintError, int, int) {
	var allErrors []rule.LintError
	allErrors = append(allErrors, rule.CheckFinalBlankLine(path, content)...)
	allErrors = append(allErrors, rule.CheckUnclosedCodeBlocks(path, content)...)
	allErrors = append(allErrors, rule.CheckEmptyAltText(path, content)...)
	if cfg.EnableHeadingLevelCheck {
		allErrors = append(allErrors, rule.CheckHeadingLevels(path, content, cfg.MinHeadingLevel)...)
	}
	if cfg.EnableDuplicateHeadingCheck {
		allErrors = append(allErrors, rule.CheckDuplicateHeadings(path, content)...)
	}
	if cfg.EnableNoMultipleBlankLinesCheck {
		allErrors = append(allErrors, rule.CheckNoMultipleBlankLines(path, content)...)
	}

	linksChecked := 0
	if cfg.EnableLinkCheck {
		links := parser.ExtractExternalLinksWithLineNumbers(content)
		// Count unique URLs
		uniqueURLs := make(map[string]bool)
		for _, link := range links {
			uniqueURLs[link.URL] = true
		}
		linksChecked = len(uniqueURLs)
		allErrors = append(allErrors, rule.CheckExternalLinks(path, content, patterns, cfg.LinkCheckTimeoutSeconds, 1000, urlCache)...)
	}

	sort.Slice(allErrors, func(i, j int) bool {
		return allErrors[i].Line < allErrors[j].Line
	})

	lineCount := strings.Count(content, "\n") + 1
	return allErrors, lineCount, linksChecked
}

func printJSONOutput(results map[string][]rule.LintError, totalFiles, totalLines, totalErrors, totalLinksChecked int, linkCheckEnabled bool, duration time.Duration) {
	output := struct {
		Files        int                         `json:"files"`
		Lines        int                         `json:"lines"`
		Errors       int                         `json:"errors"`
		LinksChecked *int                        `json:"links_checked,omitempty"`
		ElapsedMS    int64                       `json:"elapsed_ms"`
		ErrorDetail  map[string][]rule.LintError `json:"details"`
	}{
		Files:       totalFiles,
		Lines:       totalLines,
		Errors:      totalErrors,
		ElapsedMS:   duration.Milliseconds(),
		ErrorDetail: results,
	}

	if linkCheckEnabled {
		output.LinksChecked = &totalLinksChecked
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
	totalFiles, totalLines, totalErrors, totalLinksChecked int,
	linkCheckEnabled bool,
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

	if linkCheckEnabled {
		if duration < time.Second {
			fmt.Printf("%s✓%s Checked %d file(s), %d line(s), %d link(s) in %s%dms%s\n",
				green, reset, totalFiles, totalLines, totalLinksChecked, gray, duration.Milliseconds(), reset)
		} else {
			fmt.Printf("%s✓%s Checked %d file(s), %d line(s), %d link(s) in %s%.1fs%s\n",
				green, reset, totalFiles, totalLines, totalLinksChecked, gray, duration.Seconds(), reset)
		}
	} else {
		if duration < time.Second {
			fmt.Printf("%s✓%s Checked %d file(s), %d line(s) in %s%dms%s\n",
				green, reset, totalFiles, totalLines, gray, duration.Milliseconds(), reset)
		} else {
			fmt.Printf("%s✓%s Checked %d file(s), %d line(s) in %s%.1fs%s\n",
				green, reset, totalFiles, totalLines, gray, duration.Seconds(), reset)
		}
	}
}

func init() {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	rootCmd.Flags().StringVar(&configFilePath, "config", ".gomarklint.json", "path to config file (default: .gomarklint.json)")

	rootCmd.Flags().IntVar(&minHeadingLevel, "min-heading", 2, "minimum heading level to start from (default: 2)")
	rootCmd.Flags().BoolVar(&enableLinkCheck, "enable-link-check", false, "enable external link checking")
	rootCmd.Flags().BoolVar(&enableHeadingLevelCheck, "enable-heading-level-check", true, "enable heading level check")
	rootCmd.Flags().BoolVar(&enableDuplicateHeadingCheck, "enable-duplicate-heading-check", true, "enable duplicate heading check")
	rootCmd.Flags().BoolVar(&enableNoMultipleBlankLinesCheck, "enable-no-multiple-blank-lines-check", true, "enable no multiple blank lines check")
	rootCmd.Flags().StringArrayVar(&skipLinkPatterns, "skip-link-patterns", nil, "patterns of URLs to skip link checking")
	rootCmd.Flags().StringVar(&outputFormat, "output", "text", "output format: text or json")

	rootCmd.AddCommand(initCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
