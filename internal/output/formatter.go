package output

import (
	"io"
	"time"

	"github.com/shinagawa-web/gomarklint/v3/internal/rule"
)

type Formatter interface {
	Format(w io.Writer, result *Result) error
}

type Result struct {
	Files        int
	Lines        int
	Total        int
	Warnings     int
	LinksChecked *int
	Duration     time.Duration
	Details      map[string][]rule.LintError
	OrderedPaths []string
}
