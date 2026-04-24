package rule

import (
	"strings"
	"testing"
)

func TestCheckNoBareURLs(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		offset   int
		wantErrs []LintError
	}{
		{
			name:     "valid: angle bracket URL",
			content:  "Visit <https://example.com> for details.\n",
			wantErrs: nil,
		},
		{
			name:     "valid: inline Markdown link",
			content:  "Visit [Example](https://example.com) for details.\n",
			wantErrs: nil,
		},
		{
			name:     "valid: image link",
			content:  "![Alt text](https://example.com/image.png)\n",
			wantErrs: nil,
		},
		{
			name:     "valid: URL inside single-backtick inline code span",
			content:  "Use `https://example.com` as the base URL.\n",
			wantErrs: nil,
		},
		{
			name:     "valid: URL inside double-backtick inline code span",
			content:  "Use ``https://example.com`` as the base URL.\n",
			wantErrs: nil,
		},
		{
			name:     "valid: double-backtick span containing a single backtick",
			content:  "Use ``https://example.com`path`` here.\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: unclosed backtick does not suppress URL detection",
			content: "See `https://example.com for details.\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-bare-urls: bare URL found, use angle brackets or a Markdown link: https://example.com"},
			},
		},
		{
			name:    "invalid: bare URL in parentheses is still flagged",
			content: "See (https://example.com) for details.\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-bare-urls: bare URL found, use angle brackets or a Markdown link: https://example.com"},
			},
		},
		{
			name:     "valid: URL inside fenced code block",
			content:  "```\nhttps://example.com\n```\n",
			wantErrs: nil,
		},
		{
			name:     "valid: URL inside tilde fence",
			content:  "~~~\nhttps://example.com\n~~~\n",
			wantErrs: nil,
		},
		{
			name:     "valid: http angle bracket URL",
			content:  "See <http://example.com>.\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: bare https URL in text",
			content: "Visit https://example.com for details.\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-bare-urls: bare URL found, use angle brackets or a Markdown link: https://example.com"},
			},
		},
		{
			name:    "invalid: bare http URL in text",
			content: "Visit http://example.com for details.\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-bare-urls: bare URL found, use angle brackets or a Markdown link: http://example.com"},
			},
		},
		{
			name:    "invalid: bare URL at start of line",
			content: "https://example.com is the site.\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-bare-urls: bare URL found, use angle brackets or a Markdown link: https://example.com"},
			},
		},
		{
			name:    "invalid: trailing period stripped from message",
			content: "See https://example.com.\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-bare-urls: bare URL found, use angle brackets or a Markdown link: https://example.com"},
			},
		},
		{
			name:    "invalid: multiple bare URLs on one line",
			content: "See https://foo.com and https://bar.com.\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-bare-urls: bare URL found, use angle brackets or a Markdown link: https://foo.com"},
				{File: "test.md", Line: 1, Message: "no-bare-urls: bare URL found, use angle brackets or a Markdown link: https://bar.com"},
			},
		},
		{
			name:    "invalid: bare URL on second line",
			content: "## Section\n\nhttps://example.com\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "no-bare-urls: bare URL found, use angle brackets or a Markdown link: https://example.com"},
			},
		},
		{
			name:    "offset shifts line numbers",
			content: "Visit https://example.com today.\n",
			offset:  4,
			wantErrs: []LintError{
				{File: "test.md", Line: 5, Message: "no-bare-urls: bare URL found, use angle brackets or a Markdown link: https://example.com"},
			},
		},
		{
			name:     "valid: mixed — bare URL inside code, proper link in text",
			content:  "See [docs](https://example.com) and `https://internal` for reference.\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: bare URL after inline code on same line",
			content: "Run `cmd` then visit https://example.com.\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-bare-urls: bare URL found, use angle brackets or a Markdown link: https://example.com"},
			},
		},
		{
			name:     "empty file",
			content:  "",
			wantErrs: nil,
		},
		{
			name:     "valid: no URLs at all",
			content:  "## Just a heading\n\nSome plain text.\n",
			wantErrs: nil,
		},
		{
			name:     "valid: httpbin is not a URL scheme",
			content:  "Use httpbin for testing.\n",
			wantErrs: nil,
		},
		{
			name:     "valid: bare scheme with no host is ignored",
			content:  "The prefix http:// is a scheme.\n",
			wantErrs: nil,
		},
		{
			name:     "valid: angle bracket URL at column 1",
			content:  "<https://example.com>\n",
			wantErrs: nil,
		},
		{
			name:     "valid: URL inside HTML href attribute",
			content:  `<a href="https://example.com">link</a>` + "\n",
			wantErrs: nil,
		},
		{
			name:     "valid: URL inside HTML src attribute",
			content:  `<img src="https://example.com/image.gif" alt="Demo">` + "\n",
			wantErrs: nil,
		},
		{
			name:     "valid: multiple URLs inside HTML attributes on one line",
			content:  `<a href="https://example.com"><img src="https://example.com/image.gif" width="800" alt="Demo"></a>` + "\n",
			wantErrs: nil,
		},
		{
			name:     "valid: URL inside single-quoted HTML attribute",
			content:  `<a href='https://example.com'>link</a>` + "\n",
			wantErrs: nil,
		},
		{
			name:     "valid: URL inside single-line HTML comment",
			content:  "<!-- FIXME update when fixed https://example.com/ -->\n",
			wantErrs: nil,
		},
		{
			name:     "valid: URL inside multi-line HTML comment",
			content:  "<!--\nhttps://example.com\n-->\n",
			wantErrs: nil,
		},
		{
			name:     "valid: URL on same line as opening HTML comment tag (unclosed on that line)",
			content:  "<!-- https://example.com\n-->\nMore text\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: bare URL after closed HTML comment on same line",
			content: "<!-- comment --> https://example.com\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-bare-urls: bare URL found, use angle brackets or a Markdown link: https://example.com"},
			},
		},
		{
			name:    "invalid: URL surrounded by double quotes in normal prose is still flagged",
			content: "See \"https://example.com\" for details.\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-bare-urls: bare URL found, use angle brackets or a Markdown link: https://example.com"},
			},
		},
		{
			name:    "invalid: URL surrounded by single quotes in normal prose is still flagged",
			content: "See 'https://example.com' for details.\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-bare-urls: bare URL found, use angle brackets or a Markdown link: https://example.com"},
			},
		},
		{
			name:     "valid: multiple HTML comments on one line — second unclosed sets inComment",
			content:  "<!-- closed --> <!-- https://example.com\nstill inside comment\n-->\n",
			wantErrs: nil,
		},
		{
			name:    "valid: fence opener inside multi-line HTML comment is not treated as code block",
			content: "<!--\n```\nhttps://example.com\n```\n-->\nhttps://after.example.com is bare\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 6, Message: "no-bare-urls: bare URL found, use angle brackets or a Markdown link: https://after.example.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			got := CheckNoBareURLs("test.md", lines, tt.offset)

			if len(got) != len(tt.wantErrs) {
				t.Fatalf("got %d errors, want %d\ngot:  %v\nwant: %v", len(got), len(tt.wantErrs), got, tt.wantErrs)
			}
			for i := range got {
				if got[i].File != tt.wantErrs[i].File {
					t.Errorf("[%d] file: got %q, want %q", i, got[i].File, tt.wantErrs[i].File)
				}
				if got[i].Line != tt.wantErrs[i].Line {
					t.Errorf("[%d] line: got %d, want %d", i, got[i].Line, tt.wantErrs[i].Line)
				}
				if got[i].Message != tt.wantErrs[i].Message {
					t.Errorf("[%d] message: got %q, want %q", i, got[i].Message, tt.wantErrs[i].Message)
				}
			}
		})
	}
}
