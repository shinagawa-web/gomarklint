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

// withSeverity tags each error in errs with the configured severity and rule name.
func (l *Linter) withSeverity(errs []rule.LintError, ruleName string) []rule.LintError {
	sev := l.config.RuleSeverity(ruleName)
	for i := range errs {
		errs[i].Rule = ruleName
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

// maxLineLength returns the configured lineLength for the max-line-length rule.
func (l *Linter) maxLineLength() int {
	lineLength := 80
	if v, ok := l.config.RuleOptions("max-line-length")["lineLength"]; ok {
		if f, ok := v.(float64); ok && int(f) > 0 {
			lineLength = int(f)
		}
	}
	return lineLength
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

// collectLineErrors runs all non-network rule checks and returns their errors.
func (l *Linter) collectLineErrors(path string, lines []string, offset int) []rule.LintError {
	var errs []rule.LintError
	if l.config.IsEnabled("final-blank-line") {
		errs = append(errs, l.withSeverity(rule.CheckFinalBlankLine(path, lines, offset), "final-blank-line")...)
	}
	if l.config.IsEnabled("unclosed-code-block") {
		errs = append(errs, l.withSeverity(rule.CheckUnclosedCodeBlocks(path, lines, offset), "unclosed-code-block")...)
	}
	if l.config.IsEnabled("empty-alt-text") {
		errs = append(errs, l.withSeverity(rule.CheckEmptyAltText(path, lines, offset), "empty-alt-text")...)
	}
	if l.config.IsEnabled("heading-level") {
		errs = append(errs, l.withSeverity(rule.CheckHeadingLevels(path, lines, offset, l.headingMinLevel()), "heading-level")...)
	}
	if l.config.IsEnabled("fenced-code-language") {
		errs = append(errs, l.withSeverity(rule.CheckFencedCodeLanguage(path, lines, offset), "fenced-code-language")...)
	}
	if l.config.IsEnabled("duplicate-heading") {
		errs = append(errs, l.withSeverity(rule.CheckDuplicateHeadings(path, lines, offset), "duplicate-heading")...)
	}
	if l.config.IsEnabled("no-multiple-blank-lines") {
		errs = append(errs, l.withSeverity(rule.CheckNoMultipleBlankLines(path, lines, offset), "no-multiple-blank-lines")...)
	}
	if l.config.IsEnabled("no-setext-headings") {
		errs = append(errs, l.withSeverity(rule.CheckNoSetextHeadings(path, lines, offset), "no-setext-headings")...)
	}
	if l.config.IsEnabled("single-h1") {
		errs = append(errs, l.withSeverity(rule.CheckSingleH1(path, lines, offset), "single-h1")...)
	}
	if l.config.IsEnabled("blanks-around-headings") {
		errs = append(errs, l.withSeverity(rule.CheckBlanksAroundHeadings(path, lines, offset), "blanks-around-headings")...)
	}
	if l.config.IsEnabled("no-bare-urls") {
		errs = append(errs, l.withSeverity(rule.CheckNoBareURLs(path, lines, offset), "no-bare-urls")...)
	}
	if l.config.IsEnabled("no-empty-links") {
		errs = append(errs, l.withSeverity(rule.CheckNoEmptyLinks(path, lines, offset), "no-empty-links")...)
	}
	if l.config.IsEnabled("no-emphasis-as-heading") {
		errs = append(errs, l.withSeverity(rule.CheckNoEmphasisAsHeading(path, lines, offset), "no-emphasis-as-heading")...)
	}
	if l.config.IsEnabled("blanks-around-lists") {
		errs = append(errs, l.withSeverity(rule.CheckBlanksAroundLists(path, lines, offset), "blanks-around-lists")...)
	}
	if l.config.IsEnabled("max-line-length") {
		errs = append(errs, l.withSeverity(rule.CheckMaxLineLength(path, lines, offset, l.maxLineLength()), "max-line-length")...)
	}
	return errs
}

// collectErrors performs linting checks on a single file's content.
func (l *Linter) collectErrors(path string, content string) ([]rule.LintError, int, int) {
	body, offset := file.StripFrontmatter(content)
	lines := strings.Split(body, "\n")

	var disabled disabledSet
	if strings.Contains(body, "gomarklint-disable") {
		disabled = parseDisableComments(lines, offset)
	}

	allErrors := l.collectLineErrors(path, lines, offset)

	linksChecked := 0
	if l.config.IsEnabled("external-link") {
		errors, count := rule.CheckExternalLinks(path, lines, offset, l.compiledPatterns, l.externalLinkTimeout(), rule.DefaultRetryDelayMs, l.urlCache)
		allErrors = append(allErrors, l.withSeverity(errors, "external-link")...)
		linksChecked = count
	}

	if len(disabled) > 0 {
		filtered := allErrors[:0]
		for _, e := range allErrors {
			if !disabled.isDisabled(e.Line, e.Rule) {
				filtered = append(filtered, e)
			}
		}
		allErrors = filtered
	}

	sort.Slice(allErrors, func(i, j int) bool {
		return allErrors[i].Line < allErrors[j].Line
	})

	lineCount := strings.Count(content, "\n") + 1
	return allErrors, lineCount, linksChecked
}
