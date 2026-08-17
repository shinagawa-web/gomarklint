package rule

func firstNonSpaceByte(s string) byte {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
			return c
		}
	}
	return 0
}
