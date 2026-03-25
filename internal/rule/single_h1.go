package rule

import "strings"

// CheckSingleH1 checks that the document contains at most one H1 heading (ATX-style).
// Every H1 heading after the first is flagged (MD025).
// H1 headings inside fenced code blocks are ignored.
//
// Parameters:
//   - filename: the name of the file being checked (used in error reporting)
//   - lines: the Markdown content split into lines (with frontmatter already removed)
//   - offset: the line number offset due to frontmatter removal
//
// Returns:
//   - A slice of LintError, one per H1 heading after the first.
func CheckSingleH1(filename string, lines []string, offset int) []LintError {
	var errs []LintError
	inBlock := false
	fenceMarker := ""
	foundFirst := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inBlock {
			if isClosingFence(trimmed, fenceMarker) {
				inBlock = false
				fenceMarker = ""
			}
			continue
		}

		marker := openingFenceMarker(trimmed)
		if marker != "" {
			inBlock = true
			fenceMarker = marker
			continue
		}

		if !isATXH1(trimmed) {
			continue
		}

		if !foundFirst {
			foundFirst = true
			continue
		}

		errs = append(errs, LintError{
			File:    filename,
			Line:    offset + i + 1,
			Message: "Multiple H1 headings found; only one H1 is allowed per file",
		})
	}

	return errs
}

// isATXH1 reports whether line is an ATX-style H1 heading.
// A line qualifies as H1 if it starts with exactly one '#' followed by a space, tab, or end of line.
func isATXH1(line string) bool {
	if len(line) == 0 || line[0] != '#' {
		return false
	}
	if len(line) == 1 {
		return true
	}
	return line[1] != '#' && (line[1] == ' ' || line[1] == '\t')
}
