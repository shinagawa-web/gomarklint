package rule

// CheckSingleH1 flags every ATX-style H1 heading (`# ...`) after the first one in the file.
// H1 headings inside fenced code blocks are ignored.
func CheckSingleH1(filename string, lines []string, offset int) []LintError {
	var errs []LintError
	inBlock := false
	fenceMarker := ""
	foundFirst := false

	for i, line := range lines {
		// Byte-level prefilter: skip lines whose first non-space byte cannot
		// start a fence opener or an H1 heading, avoiding TrimSpace on the
		// vast majority of lines (paragraphs, list items, blank lines, etc.).
		first := firstNonSpaceByte(line)
		if inBlock {
			// Inside a fence block only a closing fence matters; the closing
			// fence character must match the opening fence character.
			if first != fenceMarker[0] {
				continue
			}
			trimmed := trimSpaceBytes(line)
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

		trimmed := trimSpaceBytes(line)

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

// firstNonSpaceByte returns the first non-whitespace byte in s, or 0 if s is
// empty or all whitespace. It is used as a cheap prefilter to skip lines that
// cannot match any rule-relevant pattern before doing heavier string work.
func firstNonSpaceByte(s string) byte {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
			return c
		}
	}
	return 0
}

// trimSpaceBytes returns s with leading and trailing ASCII whitespace removed.
// It operates purely on bytes and avoids the reflect overhead of
// strings.TrimSpace for the common ASCII-only case.
func trimSpaceBytes(s string) string {
	start := 0
	for start < len(s) {
		c := s[start]
		if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
			break
		}
		start++
	}
	end := len(s)
	for end > start {
		c := s[end-1]
		if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
			break
		}
		end--
	}
	return s[start:end]
}
