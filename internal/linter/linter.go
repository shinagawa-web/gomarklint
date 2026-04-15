package linter

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/shinagawa-web/gomarklint/v2/internal/config"
	"github.com/shinagawa-web/gomarklint/v2/internal/file"
	"github.com/shinagawa-web/gomarklint/v2/internal/rule"
)

// Linter performs linting on markdown files.
type Linter struct {
	config           config.Config
	compiledPatterns []*regexp.Regexp
	urlCache         *sync.Map
}

// Result holds the results of a linting run.
type Result struct {
	Errors            map[string][]rule.LintError // All violations (errors + warnings) per file path
	OrderedPaths      []string                    // Sorted file paths
	TotalErrors       int                         // Count of severity=error violations (used for exit code)
	TotalWarnings     int                         // Count of severity=warning violations
	TotalLines        int                         // Total number of lines checked
	TotalLinksChecked int                         // Total number of links checked
	FailedFiles       map[string]error            // Files that failed to read
}

// New creates a new Linter with the given configuration.
func New(cfg config.Config) *Linter {
	compiledPatterns := []*regexp.Regexp{}
	if cfg.IsEnabled("external-link") {
		opts := cfg.RuleOptions("external-link")
		if patterns, ok := opts["skipPatterns"]; ok {
			if arr, ok := patterns.([]interface{}); ok {
				for _, p := range arr {
					if s, ok := p.(string); ok {
						re, err := regexp.Compile(s)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Invalid skip-link-pattern: %s (error: %v)\n", s, err)
							continue
						}
						compiledPatterns = append(compiledPatterns, re)
					}
				}
			}
		}
	}

	return &Linter{
		config:           cfg,
		compiledPatterns: compiledPatterns,
		urlCache:         &sync.Map{},
	}
}

// Run performs linting on the given file paths concurrently.
func (l *Linter) Run(filePaths []string) *Result {
	// Deduplicate file paths to prevent double-counting
	uniquePaths := make(map[string]struct{})
	for _, p := range filePaths {
		uniquePaths[p] = struct{}{}
	}
	deduped := make([]string, 0, len(uniquePaths))
	for p := range uniquePaths {
		deduped = append(deduped, p)
	}

	results := map[string][]rule.LintError{}
	orderedPaths := make([]string, 0, len(deduped))
	failedFiles := map[string]error{}
	totalErrors := 0
	totalWarnings := 0
	totalLines := 0
	totalLinksChecked := 0

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, path := range deduped {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			content, err := file.ReadFile(p)
			if err != nil {
				mu.Lock()
				failedFiles[p] = err
				mu.Unlock()
				fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", p, err)
				return
			}
			errors, lineCount, linksChecked := l.collectErrors(p, content)

			mu.Lock()
			results[p] = errors
			orderedPaths = append(orderedPaths, p)
			for _, e := range errors {
				if e.Severity == "warning" {
					totalWarnings++
				} else {
					totalErrors++
				}
			}
			totalLines += lineCount
			totalLinksChecked += linksChecked
			mu.Unlock()
		}(path)
	}

	wg.Wait()

	// Sort paths to ensure consistent output order
	sort.Strings(orderedPaths)

	return &Result{
		Errors:            results,
		OrderedPaths:      orderedPaths,
		TotalErrors:       totalErrors,
		TotalWarnings:     totalWarnings,
		TotalLines:        totalLines,
		TotalLinksChecked: totalLinksChecked,
		FailedFiles:       failedFiles,
	}
}

// LintContent performs linting checks on the provided content string.
// This is useful for benchmarking and testing without file I/O overhead.
func (l *Linter) LintContent(path string, content string) ([]rule.LintError, int, int) {
	return l.collectErrors(path, content)
}

// withSeverity tags each error in errs with the configured severity for ruleName.
func (l *Linter) withSeverity(errs []rule.LintError, ruleName string) []rule.LintError {
	sev := l.config.RuleSeverity(ruleName)
	for i := range errs {
		errs[i].Severity = sev
	}
	return errs
}

// headingMinLevel returns the configured minLevel for the heading-level rule.
func (l *Linter) headingMinLevel() int {
	minLevel := 2
	if v, ok := l.config.RuleOptions("heading-level")["minLevel"]; ok {
		if f, ok := v.(float64); ok {
			minLevel = int(f)
		}
	}
	return minLevel
}

// externalLinkTimeout returns the configured timeoutSeconds for the external-link rule.
func (l *Linter) externalLinkTimeout() int {
	timeoutSeconds := 5
	if v, ok := l.config.RuleOptions("external-link")["timeoutSeconds"]; ok {
		if f, ok := v.(float64); ok && int(f) > 0 {
			timeoutSeconds = int(f)
		}
	}
	return timeoutSeconds
}

// collectErrors performs linting checks on a single file's content.
func (l *Linter) collectErrors(path string, content string) ([]rule.LintError, int, int) {
	body, offset := file.StripFrontmatter(content)
	lines := strings.Split(body, "\n")

	var allErrors []rule.LintError

	if l.config.IsEnabled("final-blank-line") {
		allErrors = append(allErrors, l.withSeverity(rule.CheckFinalBlankLine(path, lines, offset), "final-blank-line")...)
	}
	if l.config.IsEnabled("unclosed-code-block") {
		allErrors = append(allErrors, l.withSeverity(rule.CheckUnclosedCodeBlocks(path, lines, offset), "unclosed-code-block")...)
	}
	if l.config.IsEnabled("empty-alt-text") {
		allErrors = append(allErrors, l.withSeverity(rule.CheckEmptyAltText(path, lines, offset), "empty-alt-text")...)
	}
	if l.config.IsEnabled("heading-level") {
		allErrors = append(allErrors, l.withSeverity(rule.CheckHeadingLevels(path, lines, offset, l.headingMinLevel()), "heading-level")...)
	}
	if l.config.IsEnabled("fenced-code-language") {
		allErrors = append(allErrors, l.withSeverity(rule.CheckFencedCodeLanguage(path, lines, offset), "fenced-code-language")...)
	}
	if l.config.IsEnabled("duplicate-heading") {
		allErrors = append(allErrors, l.withSeverity(rule.CheckDuplicateHeadings(path, lines, offset), "duplicate-heading")...)
	}
	if l.config.IsEnabled("no-multiple-blank-lines") {
		allErrors = append(allErrors, l.withSeverity(rule.CheckNoMultipleBlankLines(path, lines, offset), "no-multiple-blank-lines")...)
	}
	if l.config.IsEnabled("no-setext-headings") {
		allErrors = append(allErrors, l.withSeverity(rule.CheckNoSetextHeadings(path, lines, offset), "no-setext-headings")...)
	}
	if l.config.IsEnabled("single-h1") {
		allErrors = append(allErrors, l.withSeverity(rule.CheckSingleH1(path, lines, offset), "single-h1")...)
	}
	if l.config.IsEnabled("blanks-around-headings") {
		allErrors = append(allErrors, l.withSeverity(rule.CheckBlanksAroundHeadings(path, lines, offset), "blanks-around-headings")...)
	}
	if l.config.IsEnabled("no-bare-urls") {
		allErrors = append(allErrors, l.withSeverity(rule.CheckNoBareURLs(path, lines, offset), "no-bare-urls")...)
	}
	if l.config.IsEnabled("no-empty-links") {
		allErrors = append(allErrors, l.withSeverity(rule.CheckNoEmptyLinks(path, lines, offset), "no-empty-links")...)
	}
	if l.config.IsEnabled("no-trailing-spaces") {
		allErrors = append(allErrors, l.withSeverity(rule.CheckNoTrailingSpaces(path, lines, offset), "no-trailing-spaces")...)
	}

	linksChecked := 0
	if l.config.IsEnabled("external-link") {
		errors, count := rule.CheckExternalLinks(path, lines, offset, l.compiledPatterns, l.externalLinkTimeout(), rule.DefaultRetryDelayMs, l.urlCache)
		allErrors = append(allErrors, l.withSeverity(errors, "external-link")...)
		linksChecked = count
	}

	sort.Slice(allErrors, func(i, j int) bool {
		return allErrors[i].Line < allErrors[j].Line
	})

	lineCount := strings.Count(content, "\n") + 1
	return allErrors, lineCount, linksChecked
}
