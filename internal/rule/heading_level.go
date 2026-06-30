package rule

import (
	"fmt"
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

// LintError represents a single lint violation detected in a Markdown file.
type LintError struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Rule     string `json:"rule,omitempty"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

// atxHeadingLevel returns the ATX heading level (1–6) if line begins with a
// valid ATX heading marker, or 0 otherwise. A valid marker is one to six '#'
// characters followed by a space, a tab, or end-of-string.
// This replaces a regex match and allocates nothing.
func atxHeadingLevel(line string) int {
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	if level == 0 || level > 6 {
		return 0
	}
	// The marker may stand alone (heading with no text) or be followed by a space/tab.
	if level == len(line) || line[level] == ' ' || line[level] == '\t' {
		return level
	}
	return 0
}

// CheckHeadingLevels analyzes the heading structure of the given Markdown content
// and reports any issues such as the first heading not starting at the specified minimum level
// or heading levels that jump more than one level (e.g., from ## to ####).
//
// Parameters:
//   - filename: the name of the file being checked (used in error reporting)
//   - ctx: the shared per-line context produced by preprocess.Scan
//   - offset: the line number offset due to frontmatter removal
//   - minLevel: the expected minimum level for the first heading (e.g., 2 for ##)
//
// Returns:
//   - A slice of LintError containing the line number and description of each detected issue.
//
// Headings inside fenced code, indented code, HTML blocks, and HTML comments are
// ignored, so they neither report nor pollute the heading-level state.
func CheckHeadingLevels(filename string, ctx *preprocess.Context, offset int, minLevel int) []LintError {
	var errs []LintError

	prevLevel := 0

	for i := 0; i < ctx.Len(); i++ {
		if inBlockContext(ctx, i) {
			continue
		}

		line := ctx.Line(i)
		// First-byte prefilter: skip lines that cannot start a heading before
		// calling strings.TrimSpace.
		if firstNonSpaceByte(line) != '#' {
			continue
		}

		trimmed := strings.TrimSpace(line)

		// Pass trimmed so that CRLF '\r' and leading spaces are already removed.
		currentLevel := atxHeadingLevel(trimmed)
		if currentLevel == 0 {
			continue
		}

		if prevLevel == 0 {
			if currentLevel != minLevel {
				errs = append(errs, LintError{
					File:    filename,
					Line:    i + 1 + offset,
					Message: fmt.Sprintf("First heading should be level %d (found level %d)", minLevel, currentLevel),
				})
			}
		} else if currentLevel > prevLevel+1 {
			errs = append(errs, LintError{
				File:    filename,
				Line:    i + 1 + offset,
				Message: fmt.Sprintf("Heading level jumped from %d to %d", prevLevel, currentLevel),
			})
		}
		prevLevel = currentLevel
	}

	return errs
}
