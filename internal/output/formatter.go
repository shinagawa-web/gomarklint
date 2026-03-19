package output

import (
	"io"
	"time"

	"github.com/shinagawa-web/gomarklint/v2/internal/rule"
)

// Formatter is the interface for formatting lint results.
type Formatter interface {
	// Format writes the formatted output to the provided writer.
	Format(w io.Writer, result *Result) error
}

// Result holds the linting results to be formatted.
type Result struct {
	Files        int                         // Total number of files checked
	Lines        int                         // Total number of lines checked
	Total        int                         // Total number of issues shown (errors + warnings, after MinSeverity filter)
	Warnings     int                         // Number of those that are warnings
	LinksChecked *int                        // Number of links checked (nil if link check disabled)
	Duration     time.Duration               // Time taken for linting
	Details      map[string][]rule.LintError // Detailed violations per file (after MinSeverity filter)
	OrderedPaths []string                    // Sorted file paths for consistent output
}
