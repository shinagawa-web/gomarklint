package preprocess

import "strings"

func countBacktickRun(s string, start int) int {
	n := 0
	for start+n < len(s) && s[start+n] == '`' {
		n++
	}
	return n
}

func findClosingBacktickRun(s string, from, delimLen int) int {
	j := from
	for j < len(s) {
		if s[j] != '`' {
			j++
			continue
		}
		runLen := countBacktickRun(s, j)
		if runLen == delimLen {
			return j
		}
		j += runLen
	}
	return -1
}

func writeBlankedCodeSpan(b *strings.Builder, line string, start int) int {
	delimLen := countBacktickRun(line, start)
	closing := findClosingBacktickRun(line, start+delimLen, delimLen)
	if closing == -1 {
		// No matching closing run — emit the backticks literally.
		for k := 0; k < delimLen; k++ {
			b.WriteByte('`')
		}
		return start + delimLen
	}
	spanLen := (closing + delimLen) - start
	for k := 0; k < spanLen; k++ {
		b.WriteByte(' ')
	}
	return closing + delimLen
}

// sanitizeInline blanks inline code spans and HTML comments with spaces
// (length-preserving). The opener that appears first wins; neither construct
// nests in the other (consistent with CommonMark).
func sanitizeInline(line string, startInComment bool) (sanitized string, endedInComment, fullyComment bool) {
	// Fast path: no backtick or comment opener — return verbatim (no allocation).
	// Returning the same string pointer lets Scan skip the sparse map entry.
	if !startInComment &&
		strings.IndexByte(line, '`') < 0 &&
		!strings.Contains(line, "<!--") {
		return line, false, false
	}

	var b strings.Builder
	b.Grow(len(line))

	inComment := startInComment
	hasComment := startInComment
	hasOther := false

	i := 0
	for i < len(line) {
		if inComment {
			if i+3 <= len(line) && line[i:i+3] == "-->" {
				b.WriteString("   ")
				i += 3
				inComment = false
			} else {
				b.WriteByte(' ')
				i++
			}
			continue
		}

		if i+4 <= len(line) && line[i:i+4] == "<!--" {
			inComment = true
			hasComment = true
			b.WriteString("    ")
			i += 4
			continue
		}

		if line[i] == '`' {
			i = writeBlankedCodeSpan(&b, line, i)
			hasOther = true
			continue
		}

		c := line[i]
		if c != ' ' && c != '\t' && c != '\r' {
			hasOther = true
		}
		b.WriteByte(c)
		i++
	}

	return b.String(), inComment, hasComment && !hasOther
}
