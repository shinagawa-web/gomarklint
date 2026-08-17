package rule

import "strings"

// no-hard-tabs strips inline code but NOT inline HTML comments (markdownlint divergence #337 Section B).
// Other rules use preprocess.Context.Sanitized which blanks both.

func countBacktickRun(s string, start int) int {
	n := 0
	for start+n < len(s) && s[start+n] == '`' {
		n++
	}
	return n
}

func stripInlineCode(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); {
		if s[i] != '`' {
			b.WriteByte(s[i])
			i++
			continue
		}

		delimLen := countBacktickRun(s, i)
		closing := -1
		j := i + delimLen
		for j < len(s) {
			if s[j] != '`' {
				j++
				continue
			}
			runLen := countBacktickRun(s, j)
			if runLen == delimLen {
				closing = j
				break
			}
			j += runLen
		}

		if closing == -1 {
			// No matching closing run — emit backticks as-is.
			for k := 0; k < delimLen; k++ {
				b.WriteByte('`')
			}
			i += delimLen
			continue
		}

		// Replace the entire span (delimiters + content) with spaces.
		spanLen := (closing + delimLen) - i
		for k := 0; k < spanLen; k++ {
			b.WriteByte(' ')
		}
		i = closing + delimLen
	}

	return b.String()
}
