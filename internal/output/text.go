package output

import (
	"fmt"
	"io"
	"time"
)

const (
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorGray  = "\033[90m"
	colorReset = "\033[0m"
)

// TextFormatter formats lint results as human-readable text with colors.
type TextFormatter struct{}

// NewTextFormatter creates a new TextFormatter.
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}

// Format implements the Formatter interface for text output.
func (f *TextFormatter) Format(w io.Writer, result *Result) error {
	if err := f.formatErrorDetails(w, result); err != nil {
		return err
	}
	if err := f.formatSummary(w, result); err != nil {
		return err
	}
	if err := f.formatStats(w, result); err != nil {
		return err
	}
	return nil
}

// formatErrorDetails prints error details for each file.
func (f *TextFormatter) formatErrorDetails(w io.Writer, result *Result) error {
	for _, path := range result.OrderedPaths {
		errors := result.Details[path]
		if len(errors) == 0 {
			continue
		}
		if _, err := fmt.Fprintf(w, "Errors in %s:\n", path); err != nil {
			return err
		}
		for _, e := range errors {
			if _, err := fmt.Fprintf(w, "  %s:%d: %s\n", e.File, e.Line, e.Message); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}
	return nil
}

// formatSummary prints the summary (errors found or no issues).
func (f *TextFormatter) formatSummary(w io.Writer, result *Result) error {
	if result.Errors > 0 {
		if _, err := fmt.Fprintf(w, "\n%s✖ %d issues found%s\n", colorRed, result.Errors, colorReset); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, "\n%s✔ No issues found%s\n", colorGreen, colorReset); err != nil {
			return err
		}
	}
	return nil
}

// formatStats prints statistics (files, lines, links, duration).
func (f *TextFormatter) formatStats(w io.Writer, result *Result) error {
	if result.LinksChecked != nil {
		return f.formatStatsWithLinks(w, result)
	}
	return f.formatStatsWithoutLinks(w, result)
}

// formatStatsWithLinks prints statistics when link checking is enabled.
func (f *TextFormatter) formatStatsWithLinks(w io.Writer, result *Result) error {
	if result.Duration < time.Second {
		_, err := fmt.Fprintf(w, "%s✓%s Checked %d file(s), %d line(s), %d link(s) in %s%dms%s\n",
			colorGreen, colorReset, result.Files, result.Lines, *result.LinksChecked, colorGray, result.Duration.Milliseconds(), colorReset)
		return err
	}
	_, err := fmt.Fprintf(w, "%s✓%s Checked %d file(s), %d line(s), %d link(s) in %s%.1fs%s\n",
		colorGreen, colorReset, result.Files, result.Lines, *result.LinksChecked, colorGray, result.Duration.Seconds(), colorReset)
	return err
}

// formatStatsWithoutLinks prints statistics when link checking is disabled.
func (f *TextFormatter) formatStatsWithoutLinks(w io.Writer, result *Result) error {
	if result.Duration < time.Second {
		_, err := fmt.Fprintf(w, "%s✓%s Checked %d file(s), %d line(s) in %s%dms%s\n",
			colorGreen, colorReset, result.Files, result.Lines, colorGray, result.Duration.Milliseconds(), colorReset)
		return err
	}
	_, err := fmt.Fprintf(w, "%s✓%s Checked %d file(s), %d line(s) in %s%.1fs%s\n",
		colorGreen, colorReset, result.Files, result.Lines, colorGray, result.Duration.Seconds(), colorReset)
	return err
}
