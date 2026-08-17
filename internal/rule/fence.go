package rule

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
