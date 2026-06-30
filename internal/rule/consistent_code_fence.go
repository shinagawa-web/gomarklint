package rule

import (
	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

// CheckConsistentCodeFence flags fenced code blocks that use a different fence
// character than expected. style must be "consistent", "backtick", or "tilde".
//
// In "consistent" mode the first fence character found in the document sets the
// expected style; every subsequent opener that differs is flagged.
// In "backtick"/"tilde" mode every opener using the wrong character is flagged.
//
// Fence openers inside indented code, HTML blocks, and HTML comments are not
// real fences and are excluded by the scanner, so they are never examined.
func CheckConsistentCodeFence(filename string, ctx *preprocess.Context, offset int, style string) []LintError {
	var errs []LintError
	var expectedCh byte // 0 until first fence seen (consistent mode)

	for _, span := range ctx.FenceSpans() {
		ch := firstNonSpaceByte(ctx.Line(span.Start))
		if err := checkFenceStyle(filename, offset+span.Start+1, ch, style, &expectedCh); err != nil {
			errs = append(errs, *err)
		}
	}

	return errs
}

// checkFenceStyle validates ch against the configured style and updates
// expectedCh in consistent mode. Returns a LintError if the fence character
// does not match, nil otherwise.
func checkFenceStyle(filename string, line int, ch byte, style string, expectedCh *byte) *LintError {
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
				Message: "consistent-code-fence: expected " + fenceCharName(*expectedCh) + " fence, got " + fenceCharName(ch) + " fence",
			}
		}
	case "backtick":
		if ch != '`' {
			return &LintError{
				File:    filename,
				Line:    line,
				Message: "consistent-code-fence: expected backtick fence, got tilde fence",
			}
		}
	case "tilde":
		if ch != '~' {
			return &LintError{
				File:    filename,
				Line:    line,
				Message: "consistent-code-fence: expected tilde fence, got backtick fence",
			}
		}
	}
	return nil
}

func fenceCharName(ch byte) string {
	if ch == '`' {
		return "backtick"
	}
	return "tilde"
}
