package rule

import "strings"

// CheckSingleH1 flags every ATX-style H1 heading (`# ...`) after the first one in the file.
// H1 headings inside fenced code blocks are ignored.
func CheckSingleH1(filename string, lines []string, offset int) []LintError {
	var errs []LintError
	inBlock := false
	fenceMarker := ""
	foundFirst := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inBlock {
			if strings.HasPrefix(trimmed, fenceMarker) && strings.TrimSpace(trimmed[len(fenceMarker):]) == "" {
				inBlock = false
				fenceMarker = ""
			}
			continue
		}

		marker := openingFenceMarker(trimmed)
		if marker != "" {
			inBlock = true
			fenceMarker = marker
			continue
		}

		if !strings.HasPrefix(trimmed, "# ") && trimmed != "#" {
			continue
		}

		if !foundFirst {
			foundFirst = true
			continue
		}

		errs = append(errs, LintError{
			File:    filename,
			Line:    offset + i + 1,
			Message: "Multiple H1 headings found; only one H1 is allowed per file",
		})
	}

	return errs
}
