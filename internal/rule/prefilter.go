package rule

// firstNonSpaceByte returns the first non-ASCII-whitespace byte in s, or 0 if
// s is empty or all ASCII whitespace. Used as a cheap prefilter so that
// strings.TrimSpace is only called on lines that can plausibly match a
// rule-relevant pattern (fence opener/closer, ATX heading).
func firstNonSpaceByte(s string) byte {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
			return c
		}
	}
	return 0
}
