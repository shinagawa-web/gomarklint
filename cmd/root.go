package cmd

import (
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

		minHeadingLevel = cfg.MinHeadingLevel
		checkLinks = cfg.CheckLinks
		skipLinkPatterns = cfg.SkipLinkPatterns

		totalLines := 0

		if len(args) == 0 {
			return fmt.Errorf("please provide a markdown file or directory")
		}
		files, err := parser.ExpandPaths(args)
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
		for _, path := range files {
			content, err := parser.ReadFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", path, err)
				continue
			}
			lines := strings.Count(content, "\n") + 1
			totalLines += lines
			allErrors := []rule.LintError{}
			allErrors = append(allErrors, rule.CheckHeadingLevels(path, content, minHeadingLevel)...)
			allErrors = append(allErrors, rule.CheckFinalBlankLine(path, content)...)
			allErrors = append(allErrors, rule.CheckUnclosedCodeBlocks(path, content)...)
			allErrors = append(allErrors, rule.CheckDuplicateHeadings(path, content)...)
			if checkLinks {
				allErrors = append(allErrors, rule.CheckExternalLinks(path, content, compiledPatterns)...)
			}
			if len(allErrors) > 0 {
				sort.Slice(allErrors, func(i, j int) bool {
					return allErrors[i].Line < allErrors[j].Line
				})
				fmt.Printf("Errors in %s:\n", path)
				for _, e := range allErrors {
					fmt.Printf("  %s:%d: %s\n", e.File, e.Line, e.Message)
				}
				fmt.Println()
				totalErrors += len(allErrors)
			}
		}

		red := "\033[31m"
		green := "\033[32m"
		gray := "\033[90m"
		reset := "\033[0m"

		if totalErrors > 0 {
			fmt.Printf("\n%s✖ %d issues found%s\n", red, totalErrors, reset)
		} else {
			fmt.Printf("\n%s✔ No issues found%s\n", green, reset)
		}

		elapsed := time.Since(start)
		if elapsed < time.Second {
			fmt.Printf("%s✓%s Checked %d file(s), %d line(s) in %s%dms%s\n",
				green, reset, len(files), totalLines, gray, elapsed.Milliseconds(), reset)
		} else {
			fmt.Printf("%s✓%s Checked %d file(s), %d line(s) in %s%.1fs%s\n",
				green, reset, len(files), totalLines, gray, elapsed.Seconds(), reset)
		}
		return nil
	},
}

func init() {
	rootCmd.Flags().StringVar(&configFilePath, "config", ".gomarklint.json", "path to config file (default: .gomarklint.json)")

	rootCmd.Flags().IntVar(&minHeadingLevel, "min-heading", 2, "minimum heading level to start from (default: 2)")
	rootCmd.Flags().BoolVar(&checkLinks, "check-links", false, "enable external link checking")
	rootCmd.Flags().StringArrayVar(&skipLinkPatterns, "skip-link-patterns", nil, "patterns of URLs to skip link checking")

	rootCmd.AddCommand(initCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
