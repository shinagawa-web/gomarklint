package preprocess

import "strings"

// openingFenceMarker returns the full fence marker (e.g. "```", "````", "~~~")
// if trimmed is an opening code fence, or an empty string otherwise. The full
// run of fence characters is captured so that closing fences of the same length
// are correctly matched. trimmed must already have its leading indentation
// removed by the caller; per CommonMark a fence is only valid at an indentation
// of less than four columns, which Scan enforces before calling this.
//
// Modeled on internal/rule/fence.go; reimplemented here so the preprocess
// package stays self-contained and has no dependency on the rule package.
func openingFenceMarker(trimmed string) string {
	if len(trimmed) < 3 {
		return ""
	}
	ch := trimmed[0]
	if (ch != '`' && ch != '~') || trimmed[1] != ch || trimmed[2] != ch {
		return ""
	}
	n := 3
	for n < len(trimmed) && trimmed[n] == ch {
		n++
	}
	return trimmed[:n]
}

// isClosingFence reports whether trimmed is a valid closing fence for the given
// opening fence marker. Per the CommonMark spec, a closing fence must:
//  1. Use the same fence character as the opener (backtick or tilde)
//  2. Have a run length >= the opener's length
//  3. Contain only optional whitespace after the fence run
func isClosingFence(trimmed, openMarker string) bool {
	if len(trimmed) == 0 || len(openMarker) == 0 {
		return false
	}
	ch := openMarker[0]
	if trimmed[0] != ch {
		return false
	}
	n := 0
	for n < len(trimmed) && trimmed[n] == ch {
		n++
	}
	if n < len(openMarker) {
		return false
	}
	return strings.TrimSpace(trimmed[n:]) == ""
}
