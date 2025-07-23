package cmd

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/shinagawa-web/gomarklint/internal/parser"
	"github.com/shinagawa-web/gomarklint/internal/rule"
	"github.com/spf13/cobra"
)

var minHeadingLevel int
var checkLinks bool
var skipLinkPatterns []string

var rootCmd = &cobra.Command{
	Use:   "gomarklint",
	Short: "A fast markdown linter written in Go",
	Long:  "gomarklint checks markdown files for common issues like heading structure, blank lines, and more.",
	RunE: func(cmd *cobra.Command, args []string) error {
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
			allErrors := []rule.LintError{}
			allErrors = append(allErrors, rule.CheckHeadingLevels(path, content, minHeadingLevel)...)
			allErrors = append(allErrors, rule.CheckFinalBlankLine(path, content)...)
			allErrors = append(allErrors, rule.CheckUnclosedCodeBlocks(path, content)...)
			if checkLinks {
				allErrors = append(allErrors, rule.CheckExternalLinks(path, content, compiledPatterns)...)
			}
			if len(allErrors) > 0 {
				fmt.Printf("Errors in %s:\n", path)
				for _, e := range allErrors {
					fmt.Printf("  %s:%d: %s\n", e.File, e.Line, e.Message)
				}
				fmt.Println()
				totalErrors += len(allErrors)
			}
		}
		if totalErrors > 0 {
			fmt.Printf("\n✖ %d issues found\n", totalErrors)
		}
		return nil
	},
}

func init() {
	rootCmd.Flags().IntVar(&minHeadingLevel, "min-heading", 2, "minimum heading level to start from (default: 2)")
	rootCmd.Flags().BoolVar(&checkLinks, "check-links", false, "enable external link checking")
	rootCmd.Flags().StringArrayVar(&skipLinkPatterns, "skip-link-patterns", nil, "patterns of URLs to skip link checking")
}

func Execute() error {
	return rootCmd.Execute()
}
