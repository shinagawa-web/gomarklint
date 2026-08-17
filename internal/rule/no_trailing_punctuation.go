package rule

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

func atxHeadingText(line string) (string, bool) {
	level := atxHeadingLevel(line)
	if level == 0 {
		return "", false
	}
	text := strings.TrimSpace(line[level:])
	// Strip optional closing ATX sequence: one or more '#' preceded by space or start.
	if n := len(text); n > 0 && text[n-1] == '#' {
		j := n - 1
		for j >= 0 && text[j] == '#' {
			j--
		}
		if j < 0 || text[j] == ' ' || text[j] == '\t' {
			text = strings.TrimRight(text[:j+1], " \t")
		}
	}
	return text, true
}

func atxLineText(first byte, line string) (string, bool) {
	if first != '#' {
		return "", false
	}
	return atxHeadingText(strings.TrimSpace(line))
}

func setextHeadingText(first byte, line, prevLine string, prevIsBlock bool) (string, bool) {
	if prevLine == "" || prevIsBlock || (first != '=' && first != '-') {
		return "", false
	}
	if !setextUnderlineRegex.MatchString(line) {
		return "", false
	}
	return strings.TrimSpace(prevLine), true
}

func isPossibleBlockMarker(b byte) bool {
	return b == '*' || b == '+' || b == '-' || b == '>' || (b >= '0' && b <= '9')
}

func noTPViolation(filename string, lineNum int, r rune) LintError {
	return LintError{
		File:    filename,
		Line:    lineNum,
		Message: fmt.Sprintf("no-trailing-punctuation: heading ends with %q", string(r)),
	}
}

func CheckNoTrailingPunctuation(filename string, ctx *preprocess.Context, offset int, punctuation string) []LintError {
	if punctuation == "" {
		return nil
	}

	var errs []LintError
	prevLine := ""       // raw previous non-blank non-block line (setext candidate)
	prevIsBlock := false // true when the previous line was a block-level element

	for i := 0; i < ctx.Len(); i++ {
		line := ctx.Line(i)
		first := firstNonSpaceByte(line)

		if first == 0 {
			prevLine = ""
			prevIsBlock = false
			continue
		}

		// A line in any code/HTML block context is not heading text, and a setext
		// underline cannot follow it.
		if inBlockContext(ctx, i) {
			prevLine = ""
			prevIsBlock = true
			continue
		}

		if text, ok := atxLineText(first, line); ok {
			if r, ok := lastRuneInSet(text, punctuation); ok {
				errs = append(errs, noTPViolation(filename, i+1+offset, r))
			}
			prevLine = ""
			prevIsBlock = true
			continue
		}

		if text, ok := setextHeadingText(first, line, prevLine, prevIsBlock); ok {
			if r, ok := lastRuneInSet(text, punctuation); ok {
				errs = append(errs, noTPViolation(filename, i+offset, r))
			}
			prevLine = ""
			prevIsBlock = true
			continue
		}

		if isPossibleBlockMarker(first) && isOtherBlockLine(line, first) {
			prevLine = ""
			prevIsBlock = true
			continue
		}

		prevLine = line
		prevIsBlock = false
	}

	return errs
}

func isOtherBlockLine(line string, first byte) bool {
	i := 0
	for i < len(line) && line[i] == ' ' {
		i++
	}
	if i > 3 || i >= len(line) || line[i] != first {
		return false
	}
	switch first {
	case '*', '+', '-', '>':
		return true
	default: // digit: requires digit+ followed by '.' or ')'
		i++
		for i < len(line) && line[i] >= '0' && line[i] <= '9' {
			i++
		}
		return i < len(line) && (line[i] == '.' || line[i] == ')')
	}
}

func lastRuneInSet(s, set string) (rune, bool) {
	if s == "" {
		return 0, false
	}
	r, _ := utf8.DecodeLastRuneInString(s)
	return r, strings.ContainsRune(set, r)
}
