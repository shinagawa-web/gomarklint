package preprocess

import "strings"

// countBacktickRun returns the number of consecutive backticks starting at
// position start in s.
func countBacktickRun(s string, start int) int {
	n := 0
	for start+n < len(s) && s[start+n] == '`' {
		n++
	}
	return n
}

// findClosingBacktickRun returns the start index of the next run of exactly
// delimLen backticks at or after from, or -1 if there is none.
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

// sanitizeInline replaces inline code spans and inline HTML comments with
// spaces so that downstream rules do not scan their contents. Length is
// preserved (each blanked byte becomes a single space) so that column positions
// in the sanitized string still line up with the original.
//
// Processing is a single left-to-right pass, so the construct that opens first
// wins: a "<!--" inside a code span is treated as code (blanked as code, not a
// comment), and a backtick inside a comment is treated as comment text. This is
// consistent with CommonMark, where neither construct nests in the other.
//
// startInComment indicates the line begins inside an HTML comment that was
// opened on a previous line. The returns are:
//   - sanitized: the line with code spans and comments blanked
//   - endedInComment: true if the line ends inside an unclosed comment
//   - fullyComment: true if the line's only non-whitespace content was comment
//     text (i.e. it is a standalone comment line, not prose with a trailing
//     comment). Always false unless a comment was actually present.
func sanitizeInline(line string, startInComment bool) (sanitized string, endedInComment, fullyComment bool) {
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

		// Opening of an inline HTML comment.
		if i+4 <= len(line) && line[i:i+4] == "<!--" {
			inComment = true
			hasComment = true
			b.WriteString("    ")
			i += 4
			continue
		}

		// Inline code span.
		if line[i] == '`' {
			delimLen := countBacktickRun(line, i)
			closing := findClosingBacktickRun(line, i+delimLen, delimLen)
			if closing == -1 {
				// No matching closing run — emit the backticks literally.
				for k := 0; k < delimLen; k++ {
					b.WriteByte('`')
				}
				hasOther = true
				i += delimLen
				continue
			}
			spanLen := (closing + delimLen) - i
			for k := 0; k < spanLen; k++ {
				b.WriteByte(' ')
			}
			hasOther = true
			i = closing + delimLen
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
