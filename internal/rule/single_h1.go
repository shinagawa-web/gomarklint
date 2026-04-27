package rule

import "strings"

// CheckSingleH1 flags every ATX-style H1 heading (`# ...`) after the first one in the file.
// H1 headings inside fenced code blocks are ignored.
func CheckSingleH1(filename string, lines []string, offset int) []LintError {
	var errs []LintError
	inBlock := false
	fenceMarker := ""
	foundFirst := false

	for i, line := range lines {
		// Byte-level prefilter: skip lines whose first non-ASCII-space byte cannot
		// start a fence opener or an H1 heading, avoiding strings.TrimSpace on the
		// vast majority of lines (paragraphs, list items, blank lines, etc.).
		first := firstNonSpaceByte(line)
		if inBlock {
			// Inside a fence block only a closing fence matters; the closing
			// fence character must match the opening fence character.
			if first != fenceMarker[0] {
				continue
			}
			trimmed := strings.TrimSpace(line)
			if IsClosingFence(trimmed, fenceMarker) {
				inBlock = false
				fenceMarker = ""
			}
			continue
		}

		// Outside a block, only '`', '~', and '#' are relevant.
		if first != '`' && first != '~' && first != '#' {
			continue
		}

		trimmed := strings.TrimSpace(line)

		if marker := openingFenceMarker(trimmed); marker != "" {
			inBlock = true
			fenceMarker = marker
			continue
		}

		if len(trimmed) == 0 || trimmed[0] != '#' {
			continue
		}
		// Must be "# ..." (H1 with space) or bare "#" (also H1).
		if len(trimmed) >= 2 && trimmed[1] != ' ' {
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

// firstNonSpaceByte returns the first non-ASCII-whitespace byte in s, or 0 if
// s is empty or all ASCII whitespace. Used as a cheap prefilter so that
// strings.TrimSpace is only called on lines that can plausibly match a
// rule-relevant pattern (fence opener/closer, ATX heading).
func firstNonSpaceByte(s string) byte {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
			return c
		}
	}
	return 0
}
