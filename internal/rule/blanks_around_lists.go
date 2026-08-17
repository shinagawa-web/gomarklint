package rule

import (
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

func isListItem(line string) bool {
	s := strings.TrimLeft(line, " \t")
	if len(s) < 2 {
		return false
	}
	return isUnorderedListItem(s) || isOrderedListItem(s)
}

func isUnorderedListItem(s string) bool {
	if s[0] != '-' && s[0] != '*' && s[0] != '+' {
		return false
	}
	return s[1] == ' ' || s[1] == '\t'
}

func isOrderedListItem(s string) bool {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 || i >= len(s) {
		return false
	}
	if s[i] != '.' && s[i] != ')' {
		return false
	}
	if i+1 >= len(s) {
		return false
	}
	return s[i+1] == ' ' || s[i+1] == '\t'
}

func CheckBlanksAroundLists(filename string, ctx *preprocess.Context, offset int) []LintError {
	var errs []LintError
	// prevBlank and prevWasListItem replace the TrimSpace look-behind on
	// lines[i-1] for every list item encountered.
	// prevBlank starts as true to model the pre-file boundary as blank; the
	// first-line exemption is enforced by the i > 0 guard below.
	prevBlank := true
	prevWasListItem := false
	prevLineNum := 0 // 1-indexed line number of the previous list item

	for i := 0; i < ctx.Len(); i++ {
		line := ctx.Line(i)
		trimmed := strings.TrimSpace(line)
		isBlank := trimmed == ""
		isList := isListItem(line)

		// Check "end of block" before the block skip so a list item immediately
		// followed by a code/HTML block opener is still flagged (lesson from PR-4).
		if prevWasListItem && !isBlank && !isList {
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + prevLineNum + 1,
				Message: "blanks-around-lists: list must be followed by a blank line",
			})
		}

		if inBlockContext(ctx, i) {
			prevBlank = false
			prevWasListItem = false
			continue
		}

		if isList {
			// Check "start of block": first item of a block not preceded by blank.
			if i > 0 && !prevBlank && !prevWasListItem {
				errs = append(errs, LintError{
					File:    filename,
					Line:    offset + i + 1,
					Message: "blanks-around-lists: list must be preceded by a blank line",
				})
			}
			prevWasListItem = true
			prevLineNum = i + 1
		} else {
			prevWasListItem = false
		}

		prevBlank = isBlank
	}

	return errs
}
