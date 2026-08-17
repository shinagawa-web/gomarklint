package rule

import (
	"fmt"
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

type LintError struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Rule     string `json:"rule,omitempty"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

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
