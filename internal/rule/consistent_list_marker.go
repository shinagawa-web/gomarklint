package rule

import (
	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

func CheckConsistentListMarker(filename string, ctx *preprocess.Context, offset int, style string) []LintError {
	var errs []LintError
	var expectedCh byte // 0 until first list item seen (consistent mode)

	for i := 0; i < ctx.Len(); i++ {
		if inBlockContext(ctx, i) {
			continue
		}

		line := ctx.Line(i)
		first := firstNonSpaceByte(line)
		if first != '-' && first != '*' && first != '+' {
			continue
		}

		ch, ok := listItemMarker(line)
		if !ok {
			continue
		}

		if err := checkListMarkerStyle(filename, offset+i+1, ch, style, &expectedCh); err != nil {
			errs = append(errs, *err)
		}
	}

	return errs
}

func listItemMarker(line string) (byte, bool) {
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	if i >= len(line) {
		return 0, false
	}
	ch := line[i]
	i++
	// must be followed by at least one space or tab
	if i >= len(line) || (line[i] != ' ' && line[i] != '\t') {
		return 0, false
	}
	// skip all whitespace after the marker
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	// must have non-whitespace content (\r counts as whitespace here)
	if i >= len(line) || line[i] == '\r' || line[i] == '\n' {
		return 0, false
	}
	return ch, true
}

func checkListMarkerStyle(filename string, line int, ch byte, style string, expectedCh *byte) *LintError {
	switch style {
	case "consistent":
		if *expectedCh == 0 {
			*expectedCh = ch
			return nil
		}
		if ch != *expectedCh {
			return &LintError{
				File:    filename,
				Line:    line,
				Message: "consistent-list-marker: expected " + listMarkerName(*expectedCh) + " marker, got " + listMarkerName(ch) + " marker",
			}
		}
	case "dash":
		if ch != '-' {
			return &LintError{
				File:    filename,
				Line:    line,
				Message: "consistent-list-marker: expected dash marker, got " + listMarkerName(ch) + " marker",
			}
		}
	case "asterisk":
		if ch != '*' {
			return &LintError{
				File:    filename,
				Line:    line,
				Message: "consistent-list-marker: expected asterisk marker, got " + listMarkerName(ch) + " marker",
			}
		}
	case "plus":
		if ch != '+' {
			return &LintError{
				File:    filename,
				Line:    line,
				Message: "consistent-list-marker: expected plus marker, got " + listMarkerName(ch) + " marker",
			}
		}
	}
	return nil
}

func listMarkerName(ch byte) string {
	switch ch {
	case '-':
		return "dash"
	case '*':
		return "asterisk"
	default:
		return "plus"
	}
}
