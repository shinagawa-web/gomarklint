package rule

func CheckFinalBlankLine(filename string, lines []string, offset int) []LintError {
	// A frontmatter-only file has an empty body (lines==[""]). Offset > 0 means
	// frontmatter was stripped; there is no body to enforce the rule on.
	if len(lines) == 1 && lines[0] == "" && offset > 0 {
		return nil
	}
	if len(lines) >= 2 && lines[len(lines)-1] == "" {
		return nil
	}
	return []LintError{{
		File:    filename,
		Line:    len(lines) + offset,
		Message: "Missing final blank line",
	}}
}
