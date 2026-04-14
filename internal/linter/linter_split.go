package linter

import (
	"strings"

	"github.com/shinagawa-web/gomarklint/v2/internal/file"
	"github.com/shinagawa-web/gomarklint/v2/internal/rule"
)

// collectErrorsSplit is an alternative to collectErrors that splits rule
// execution into two separate functions. Used for benchmark investigation
// to test whether splitting a large function into smaller ones avoids
// compiler optimization regressions on amd64.
func (l *Linter) collectErrorsSplit(path string, content string) ([]rule.LintError, int, int) {
	body, offset := file.StripFrontmatter(content)
	lines := strings.Split(body, "\n")

	var allErrors []rule.LintError
	allErrors = l.collectGroupA(allErrors, path, body, lines, offset)
	allErrors = l.collectGroupB(allErrors, path, lines, offset)

	linksChecked := 0
	if l.config.IsEnabled("external-link") {
		errors, count := rule.CheckExternalLinks(path, lines, offset, l.compiledPatterns, l.externalLinkTimeout(), rule.DefaultRetryDelayMs, l.urlCache)
		allErrors = append(allErrors, l.withSeverity(errors, "external-link")...)
		linksChecked = count
	}

	lineCount := strings.Count(content, "\n") + 1
	return allErrors, lineCount, linksChecked
}

//go:noinline
func (l *Linter) collectGroupA(allErrors []rule.LintError, path, body string, lines []string, offset int) []rule.LintError {
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
	if l.config.IsEnabled("no-trailing-spaces") {
		allErrors = append(allErrors, l.withSeverity(rule.CheckNoTrailingSpaces(path, body, lines, offset), "no-trailing-spaces")...)
	}
	return allErrors
}

//go:noinline
func (l *Linter) collectGroupB(allErrors []rule.LintError, path string, lines []string, offset int) []rule.LintError {
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
	return allErrors
}

// LintContentSplit is like LintContent but uses the split approach.
func (l *Linter) LintContentSplit(path string, content string) ([]rule.LintError, int, int) {
	return l.collectErrorsSplit(path, content)
}
