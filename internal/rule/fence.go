package rule

import "strings"

// IsClosingFence reports whether trimmed is a valid closing fence for the given
// opening fence marker. Per the CommonMark spec, a closing fence must:
//  1. Use the same fence character as the opener (backtick or tilde)
//  2. Have a run length >= the opener's length
//  3. Contain only optional whitespace after the fence run
func IsClosingFence(trimmed, openMarker string) bool {
	if len(trimmed) == 0 || len(openMarker) == 0 {
		return false
	}
	ch := openMarker[0]
	if trimmed[0] != ch {
		return false
	}
	// Count the run of fence characters
	n := 0
	for n < len(trimmed) && trimmed[n] == ch {
		n++
	}
	if n < len(openMarker) {
		return false
	}
	return strings.TrimSpace(trimmed[n:]) == ""
}
