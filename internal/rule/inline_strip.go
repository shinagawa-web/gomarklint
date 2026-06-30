package rule

import "strings"

// This file holds inline-code-stripping helpers shared by several rules that
// have not yet been migrated to the preprocess context (#337 Phase 3):
// no-empty-links, link-fragments, consistent-emphasis-style, and no-hard-tabs.
// They duplicate logic now centralized in internal/preprocess (see
// Context.Sanitized) and are removed once their last consumer migrates.

// countBacktickRun returns the number of consecutive backticks starting at
// position start in s.
func countBacktickRun(s string, start int) int {
	n := 0
	for start+n < len(s) && s[start+n] == '`' {
		n++
	}
	return n
}

// stripInlineCode replaces content inside backtick spans (including the
// delimiters) with spaces so that URLs within inline code are not scanned.
// Handles single- and multi-backtick code spans (e.g. `code`) per CommonMark:
// a code span opens with a run of N backticks and closes with the next run of
// exactly N backticks, so a longer run can contain shorter ones.
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
