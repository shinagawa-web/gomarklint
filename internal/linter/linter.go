package linter

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/shinagawa-web/gomarklint/v3/internal/config"
	"github.com/shinagawa-web/gomarklint/v3/internal/file"
	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
	"github.com/shinagawa-web/gomarklint/v3/internal/rule"
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
// Returns an error if any rule option value is invalid.
func New(cfg config.Config) (*Linter, error) {
	if err := validateStyleOption(cfg, "consistent-code-fence", "style", []string{"consistent", "backtick", "tilde"}); err != nil {
		return nil, err
	}
	if err := validateStyleOption(cfg, "consistent-emphasis-style", "style", []string{"consistent", "asterisk", "underscore"}); err != nil {
		return nil, err
	}
	if err := validateStyleOption(cfg, "consistent-list-marker", "style", []string{"consistent", "dash", "asterisk", "plus"}); err != nil {
		return nil, err
	}

	if err := validateExternalLinkIntOption(cfg, "maxConcurrency", 1, rule.MaxConcurrencyLimit); err != nil {
		return nil, err
	}
	if err := validateExternalLinkIntOption(cfg, "maxRetries", 0, rule.MaxRetriesLimit); err != nil {
		return nil, err
	}
	if err := validateExternalLinkIntOption(cfg, "perHostConcurrency", 1, rule.MaxPerHostConcurrencyLimit); err != nil {
		return nil, err
	}
	if err := validatePerHostIntervalMs(cfg); err != nil {
		return nil, err
	}

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
	}, nil
}

// validateStyleOption checks that the named option for a rule, if present and non-empty,
// is one of the valid values. Returns a descriptive error if not.
func validateStyleOption(cfg config.Config, ruleName, optKey string, valid []string) error {
	opts := cfg.RuleOptions(ruleName)
	raw, exists := opts[optKey]
	if !exists {
		return nil
	}
	val, ok := raw.(string)
	if !ok {
		return fmt.Errorf("gomarklint: invalid value for %s.%s: expected string, got %T (%#v) (valid values: %s)", ruleName, optKey, raw, raw, strings.Join(valid, ", "))
	}
	if val == "" {
		return nil
	}
	for _, v := range valid {
		if val == v {
			return nil
		}
	}
	return fmt.Errorf("gomarklint: invalid value %q for %s.%s (valid values: %s)", val, ruleName, optKey, strings.Join(valid, ", "))
}

// validatePerHostIntervalMs checks that perHostIntervalMs is either 0 (disabled) or within
// [MinPerHostIntervalMs, MaxPerHostIntervalMsLimit]. Values between 1 and 999 are rejected.
func validatePerHostIntervalMs(cfg config.Config) error {
	raw, exists := cfg.RuleOptions("external-link")["perHostIntervalMs"]
	if !exists {
		return nil
	}
	f, ok := raw.(float64)
	if !ok {
		return fmt.Errorf("gomarklint: invalid value for external-link.perHostIntervalMs: expected integer, got %T (%#v)", raw, raw)
	}
	v := int(f)
	if v == 0 {
		return nil
	}
	if v < rule.MinPerHostIntervalMs || v > rule.MaxPerHostIntervalMsLimit {
		return fmt.Errorf("gomarklint: external-link.perHostIntervalMs must be 0 (disabled) or between %d and %d, got %d", rule.MinPerHostIntervalMs, rule.MaxPerHostIntervalMsLimit, v)
	}
	return nil
}

// validateExternalLinkIntOption checks that a numeric external-link option, if present,
// is within [minVal, maxVal]. Returns a descriptive error if not.
func validateExternalLinkIntOption(cfg config.Config, optKey string, minVal, maxVal int) error {
	raw, exists := cfg.RuleOptions("external-link")[optKey]
	if !exists {
		return nil
	}
	f, ok := raw.(float64)
	if !ok {
		return fmt.Errorf("gomarklint: invalid value for external-link.%s: expected integer, got %T (%#v)", optKey, raw, raw)
	}
	v := int(f)
	if v < minVal || v > maxVal {
		return fmt.Errorf("gomarklint: external-link.%s must be between %d and %d, got %d", optKey, minVal, maxVal, v)
	}
	return nil
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

// noTrailingPunctuation returns the configured punctuation string for the no-trailing-punctuation rule.
func (l *Linter) noTrailingPunctuation() string {
	if v, ok := l.config.RuleOptions("no-trailing-punctuation")["punctuation"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return config.DefaultNoTrailingPunctuation
}

// consistentCodeFenceStyle returns the configured style for the consistent-code-fence rule.
func (l *Linter) consistentCodeFenceStyle() string {
	style, _ := l.config.RuleOptions("consistent-code-fence")["style"].(string)
	if style == "" {
		return "consistent"
	}
	return style
}

// consistentEmphasisStyle returns the configured style for the consistent-emphasis-style rule.
func (l *Linter) consistentEmphasisStyle() string {
	style, _ := l.config.RuleOptions("consistent-emphasis-style")["style"].(string)
	if style == "" {
		return "consistent"
	}
	return style
}

// consistentListMarkerStyle returns the configured style for the consistent-list-marker rule.
func (l *Linter) consistentListMarkerStyle() string {
	style, _ := l.config.RuleOptions("consistent-list-marker")["style"].(string)
	if style == "" {
		return "consistent"
	}
	return style
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

// externalLinkMaxConcurrency returns the configured maxConcurrency for the external-link rule.
func (l *Linter) externalLinkMaxConcurrency() int {
	if v, ok := l.config.RuleOptions("external-link")["maxConcurrency"]; ok {
		if f, ok := v.(float64); ok && int(f) > 0 {
			return int(f)
		}
	}
	return rule.DefaultMaxConcurrency
}

// externalLinkMaxRetries returns the configured maxRetries for the external-link rule.
func (l *Linter) externalLinkMaxRetries() int {
	if v, ok := l.config.RuleOptions("external-link")["maxRetries"]; ok {
		if f, ok := v.(float64); ok && int(f) >= 0 {
			return int(f)
		}
	}
	return rule.DefaultMaxRetries
}

// externalLinkPerHostConcurrency returns the configured perHostConcurrency for the external-link rule.
func (l *Linter) externalLinkPerHostConcurrency() int {
	if v, ok := l.config.RuleOptions("external-link")["perHostConcurrency"]; ok {
		if f, ok := v.(float64); ok && int(f) > 0 {
			return int(f)
		}
	}
	return rule.DefaultPerHostConcurrency
}

// externalLinkPerHostIntervalMs returns the configured perHostIntervalMs for the external-link rule.
func (l *Linter) externalLinkPerHostIntervalMs() int {
	if v, ok := l.config.RuleOptions("external-link")["perHostIntervalMs"]; ok {
		if f, ok := v.(float64); ok && int(f) >= 0 {
			return int(f)
		}
	}
	return rule.DefaultPerHostIntervalMs
}

// externalLinkAllowedStatuses returns the configured allowedStatuses for the external-link rule.
func (l *Linter) externalLinkAllowedStatuses() []int {
	raw, _ := l.config.RuleOptions("external-link")["allowedStatuses"].([]interface{})
	statuses := make([]int, 0, len(raw))
	for _, item := range raw {
		if f, ok := item.(float64); ok {
			statuses = append(statuses, int(f))
		}
	}
	return statuses
}

// simpleRules lists rules whose check function takes only (path, lines, offset)
// and have not yet been migrated to the preprocess context. Rules that require
// additional options, or that consume the shared *preprocess.Context, are
// handled separately below.
var simpleRules = []struct {
	name string
	fn   func(string, []string, int) []rule.LintError
}{
	{"final-blank-line", rule.CheckFinalBlankLine},
	{"no-multiple-blank-lines", rule.CheckNoMultipleBlankLines},
	{"blanks-around-lists", rule.CheckBlanksAroundLists},
	{"no-hard-tabs", rule.CheckNoHardTabs},
}

// contextRules lists rules migrated to consume the shared *preprocess.Context
// (issue #337). They take (path, ctx, offset); rules that also need extra
// options are still handled separately below.
var contextRules = []struct {
	name string
	fn   func(string, *preprocess.Context, int) []rule.LintError
}{
	{"no-bare-urls", rule.CheckNoBareURLs},
	{"single-h1", rule.CheckSingleH1},
	{"duplicate-heading", rule.CheckDuplicateHeadings},
	{"no-setext-headings", rule.CheckNoSetextHeadings},
	{"blanks-around-headings", rule.CheckBlanksAroundHeadings},
	{"no-emphasis-as-heading", rule.CheckNoEmphasisAsHeading},
	{"unclosed-code-block", rule.CheckUnclosedCodeBlocks},
	{"fenced-code-language", rule.CheckFencedCodeLanguage},
	{"blanks-around-fences", rule.CheckBlanksAroundFences},
	{"empty-alt-text", rule.CheckEmptyAltText},
	{"no-empty-links", rule.CheckNoEmptyLinks},
}

// collectLineErrors runs all non-network rule checks and returns their errors.
// ctx is the shared per-line context produced once by preprocess.Scan; rules
// that have migrated to consume it receive ctx, while the rest continue to read
// the raw lines slice (which Scan borrows, so no copy is made).
func (l *Linter) collectLineErrors(path string, lines []string, ctx *preprocess.Context, offset int) []rule.LintError {
	var errs []rule.LintError

	for _, r := range simpleRules {
		if l.config.IsEnabled(r.name) {
			errs = append(errs, l.withSeverity(r.fn(path, lines, offset), r.name)...)
		}
	}

	for _, r := range contextRules {
		if l.config.IsEnabled(r.name) {
			errs = append(errs, l.withSeverity(r.fn(path, ctx, offset), r.name)...)
		}
	}

	if l.config.IsEnabled("heading-level") {
		errs = append(errs, l.withSeverity(rule.CheckHeadingLevels(path, ctx, offset, l.headingMinLevel()), "heading-level")...)
	}
	if l.config.IsEnabled("consistent-code-fence") {
		errs = append(errs, l.withSeverity(rule.CheckConsistentCodeFence(path, ctx, offset, l.consistentCodeFenceStyle()), "consistent-code-fence")...)
	}
	if l.config.IsEnabled("consistent-emphasis-style") {
		errs = append(errs, l.withSeverity(rule.CheckConsistentEmphasisStyle(path, lines, offset, l.consistentEmphasisStyle()), "consistent-emphasis-style")...)
	}
	if l.config.IsEnabled("consistent-list-marker") {
		errs = append(errs, l.withSeverity(rule.CheckConsistentListMarker(path, lines, offset, l.consistentListMarkerStyle()), "consistent-list-marker")...)
	}
	if l.config.IsEnabled("max-line-length") {
		errs = append(errs, l.withSeverity(rule.CheckMaxLineLength(path, lines, offset, l.maxLineLength()), "max-line-length")...)
	}
	if l.config.IsEnabled("no-trailing-punctuation") {
		errs = append(errs, l.withSeverity(rule.CheckNoTrailingPunctuation(path, lines, offset, l.noTrailingPunctuation()), "no-trailing-punctuation")...)
	}
	if l.config.IsEnabled("link-fragments") {
		errs = append(errs, l.withSeverity(rule.CheckLinkFragments(path, ctx, offset, l.config.RuleOptions("link-fragments")), "link-fragments")...)
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

	ctx := preprocess.Scan(lines)

	allErrors := l.collectLineErrors(path, lines, ctx, offset)

	linksChecked := 0
	if l.config.IsEnabled("external-link") {
		errors, count := rule.CheckExternalLinks(path, ctx, offset, l.compiledPatterns, l.externalLinkTimeout(), rule.DefaultRetryDelayMs, l.externalLinkMaxConcurrency(), l.externalLinkMaxRetries(), l.externalLinkAllowedStatuses(), l.urlCache, l.externalLinkPerHostConcurrency(), l.externalLinkPerHostIntervalMs())
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
