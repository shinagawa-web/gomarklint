package rule

import "strings"

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

// isClosingFence reports whether trimmed is a valid closing fence for the given fenceMarker.
// Per CommonMark spec, a closing fence must consist of at least as many fence characters as
// the opening fence, followed only by optional whitespace.
func isClosingFence(trimmed, fenceMarker string) bool {
	if len(trimmed) == 0 || trimmed[0] != fenceMarker[0] {
		return false
	}
	j := 0
	for j < len(trimmed) && trimmed[j] == fenceMarker[0] {
		j++
	}
	return j >= len(fenceMarker) && strings.TrimSpace(trimmed[j:]) == ""
}
