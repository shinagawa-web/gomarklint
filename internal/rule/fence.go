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

// openingFenceMarker returns the full fence marker (e.g. "```", "````", "~~~") if the line
// is an opening fence, or an empty string otherwise. The full run of fence characters is
// captured so that closing fences of the same length are correctly matched.
// Returns a substring of trimmed to avoid heap allocation.
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
