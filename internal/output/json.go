package output

import (
	"encoding/json"
	"io"

	"github.com/shinagawa-web/gomarklint/internal/rule"
)

// JSONFormatter formats lint results as JSON.
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSONFormatter.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// Format implements the Formatter interface for JSON output.
func (f *JSONFormatter) Format(w io.Writer, result *Result) error {
	output := struct {
		Files        int                         `json:"files"`
		Lines        int                         `json:"lines"`
		Errors       int                         `json:"errors"`
		LinksChecked *int                        `json:"links_checked,omitempty"`
		ElapsedMS    int64                       `json:"elapsed_ms"`
		ErrorDetail  map[string][]rule.LintError `json:"details"`
	}{
		Files:       result.Files,
		Lines:       result.Lines,
		Errors:      result.Errors,
		ElapsedMS:   result.Duration.Milliseconds(),
		ErrorDetail: result.Details,
	}

	if result.LinksChecked != nil {
		output.LinksChecked = result.LinksChecked
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(output); err != nil {
		return err
	}

	return nil
}
