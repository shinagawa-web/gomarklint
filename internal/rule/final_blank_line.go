package rule

// CheckFinalBlankLine checks whether the Markdown content ends with a blank line.
// This is a common requirement in many Markdown style guides.
//
// Parameters:
//   - filename: the name of the file being checked (used in error reporting)
//   - lines: the Markdown content split into lines (with frontmatter already removed)
//   - offset: the line number offset due to frontmatter removal
//
// Returns:
//   - A slice of LintError with one entry if the final blank line is missing.
func CheckFinalBlankLine(filename string, lines []string, offset int) []LintError {
	var errs []LintError
	if len(lines) < 2 || lines[len(lines)-1] != "" {
		errs = append(errs, LintError{
			File:    filename,
			Line:    len(lines) + offset,
			Message: "Missing final blank line",
		})
	}

	return errs
}
