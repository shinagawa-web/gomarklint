package preprocess

import "strings"

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

// isClosingFence reports whether trimmed closes openMarker: same character,
// run length >= opener, only optional whitespace after the run.
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
