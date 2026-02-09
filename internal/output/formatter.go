package output

import (
	"io"
	"time"

	"github.com/shinagawa-web/gomarklint/internal/rule"
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
	Errors       int                         // Total number of errors found
	LinksChecked *int                        // Number of links checked (nil if link check disabled)
	Duration     time.Duration               // Time taken for linting
	Details      map[string][]rule.LintError // Detailed errors per file
	OrderedPaths []string                    // Sorted file paths for consistent output
}
