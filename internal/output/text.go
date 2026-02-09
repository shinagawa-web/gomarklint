package output

import (
	"fmt"
	"io"
	"time"
)

// TextFormatter formats lint results as human-readable text with colors.
type TextFormatter struct{}

// NewTextFormatter creates a new TextFormatter.
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}

// Format implements the Formatter interface for text output.
func (f *TextFormatter) Format(w io.Writer, result *Result) error {
	red := "\033[31m"
	green := "\033[32m"
	gray := "\033[90m"
	reset := "\033[0m"

	// Print errors for each file
	for _, path := range result.OrderedPaths {
		errors := result.Details[path]
		if len(errors) == 0 {
			continue
		}
		fmt.Fprintf(w, "Errors in %s:\n", path)
		for _, e := range errors {
			fmt.Fprintf(w, "  %s:%d: %s\n", e.File, e.Line, e.Message)
		}
		fmt.Fprintln(w)
	}

	// Print summary
	if result.Errors > 0 {
		fmt.Fprintf(w, "\n%s✖ %d issues found%s\n", red, result.Errors, reset)
	} else {
		fmt.Fprintf(w, "\n%s✔ No issues found%s\n", green, reset)
	}

	// Print statistics
	if result.LinksChecked != nil {
		// Link check enabled
		if result.Duration < time.Second {
			fmt.Fprintf(w, "%s✓%s Checked %d file(s), %d line(s), %d link(s) in %s%dms%s\n",
				green, reset, result.Files, result.Lines, *result.LinksChecked, gray, result.Duration.Milliseconds(), reset)
		} else {
			fmt.Fprintf(w, "%s✓%s Checked %d file(s), %d line(s), %d link(s) in %s%.1fs%s\n",
				green, reset, result.Files, result.Lines, *result.LinksChecked, gray, result.Duration.Seconds(), reset)
		}
	} else {
		// Link check disabled
		if result.Duration < time.Second {
			fmt.Fprintf(w, "%s✓%s Checked %d file(s), %d line(s) in %s%dms%s\n",
				green, reset, result.Files, result.Lines, gray, result.Duration.Milliseconds(), reset)
		} else {
			fmt.Fprintf(w, "%s✓%s Checked %d file(s), %d line(s) in %s%.1fs%s\n",
				green, reset, result.Files, result.Lines, gray, result.Duration.Seconds(), reset)
		}
	}

	return nil
}
