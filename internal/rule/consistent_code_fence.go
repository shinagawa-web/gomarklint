package rule

import (
	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

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
