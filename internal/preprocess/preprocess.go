// Package preprocess provides a single shared pre-pass over a Markdown file's
// lines, classifying each line by the context it lives in (fenced code,
// indented code, HTML block, HTML comment) and exposing an inline-sanitized
// view with inline code spans and inline HTML comments blanked out.
//
// It exists to replace the per-rule "skip code blocks / HTML" machinery that
// each rule currently re-implements, which the audit in issue #337 found to be
// inconsistent and incomplete. Rules consume the result instead of re-deriving
// context, so the three systematically missed contexts — indented code blocks,
// HTML blocks, and HTML comments — are handled once, the same way, everywhere.
//
// The classification follows CommonMark's block structure but is intentionally
// not a full parser: it does not model container blocks (lists, block quotes),
// so indentation is measured from the start of the line. See the note in
// indented_code.go for the consequences.
//
// The result is stored compactly: context flags are packed one byte per line and
// the sanitized text is materialized only for the lines that actually differ
// from the original (those carrying an inline code span or comment). The
// original lines are borrowed, not copied. This keeps the per-line overhead near
// one byte rather than two string headers plus four bools, which matters because
// the linter runs Scan over every file. Access the result through the Context
// methods rather than touching the fields.
package preprocess

import "strings"

// Context is the result of Scan: the per-line Markdown classification of a file,
// queried by line index through its methods.
//
// The four context predicates (InFencedCode, InIndentedCode, InHTMLBlock,
// InHTMLComment) describe structural block membership and are mutually
// exclusive: a line is in at most one of them. They are exposed individually
// rather than via a single "skippable" convenience predicate because different
// rules skip different subsets — for example, rules implementing the markdownlint
// divergences in issue #337 (max-line-length, no-hard-tabs) deliberately scan
// inside fenced code blocks.
//
// Sanitized returns an inline-level transformation only: inline code spans and
// inline HTML comments are replaced by spaces (length-preserving), so a rule can
// scan it without matching content that lives in those spans. Block-level code
// (fenced and indented) is NOT blanked — it is returned verbatim so that rules
// which want to inspect code-block contents still can; use the predicates to skip
// block code instead. For lines reported by InHTMLComment, Sanitized is fully
// blank because the entire line was comment text.
type Context struct {
	// lines is the borrowed input slice. The caller must not mutate it for the
	// lifetime of the Context, since Line and Sanitized read from it directly.
	lines []string
	// flags holds the packed context bits for each line, parallel to lines.
	flags []uint8
	// sanitized holds the inline-sanitized text only for the lines that differ
	// from their original (those with an inline code span or comment). Lines
	// absent from the map are their own sanitized form. nil until the first such
	// line is seen.
	sanitized map[int]string
	// fences records one span per fenced code block, in document order. It is
	// the structural view that InFencedCode (a per-line membership flag) cannot
	// give: two back-to-back fences with no blank line between them are distinct
	// spans here even though every line is InFencedCode.
	fences []FenceSpan
}

// FenceSpan is the line range of one fenced code block. Start is the 0-based
// opening-fence line; End is the 0-based closing-fence line, or -1 when the
// fence is never closed (it then runs to end of file).
type FenceSpan struct {
	Start int
	End   int
}

// Context flag bits, packed one set per line in Context.flags.
const (
	flagFencedCode uint8 = 1 << iota
	flagIndentedCode
	flagHTMLBlock
	flagHTMLComment
)

// Len returns the number of lines classified (equal to len(input)).
func (c *Context) Len() int { return len(c.lines) }

// Line returns the original text of line i, exactly as it appeared in the input.
func (c *Context) Line(i int) string { return c.lines[i] }

// Sanitized returns line i with inline code spans and inline HTML comments
// replaced by spaces. For block-code and HTML-block lines it equals Line(i); for
// fully-commented lines it is all whitespace.
func (c *Context) Sanitized(i int) string {
	if s, ok := c.sanitized[i]; ok {
		return s
	}
	return c.lines[i]
}

// InFencedCode reports whether line i is part of a fenced code block, including
// the opening and closing fence lines. If a fence is never closed, every line
// from the opener to end of file is flagged, which prevents downstream rules from
// mispairing fences inside the unclosed region.
func (c *Context) InFencedCode(i int) bool { return c.flags[i]&flagFencedCode != 0 }

// InIndentedCode reports whether line i is inside an indented code block
// (CommonMark §4.4). See the limitation noted in indented_code.go.
func (c *Context) InIndentedCode(i int) bool { return c.flags[i]&flagIndentedCode != 0 }

// InHTMLBlock reports whether line i is inside an HTML block of CommonMark types
// 1 and 3–7 (e.g. <div>…</div>). HTML comments are reported via InHTMLComment,
// not this predicate.
func (c *Context) InHTMLBlock(i int) bool { return c.flags[i]&flagHTMLBlock != 0 }

// InHTMLComment reports whether line i is a line whose entire content is an HTML
// comment — a standalone <!-- … --> line or a continuation line of a multi-line
// comment. A prose line with a trailing inline comment is not reported here; its
// comment is blanked in Sanitized instead.
func (c *Context) InHTMLComment(i int) bool { return c.flags[i]&flagHTMLComment != 0 }

// FenceSpans returns the fenced code blocks in document order. Unlike the
// InFencedCode flag, this distinguishes adjacent blocks and identifies each
// block's opening and closing lines (End == -1 for an unclosed block). It is the
// structural source of truth for fence rules.
func (c *Context) FenceSpans() []FenceSpan { return c.fences }

// Scan classifies every line of a Markdown file in a single pass and returns a
// Context to query by line index. The input slice is borrowed, not copied, and
// must not be mutated while the Context is in use. The input is expected to have
// already had YAML front matter stripped by the caller (the linter does this
// centrally), so Scan does not handle front matter.
func Scan(lines []string) *Context {
	c := &Context{
		lines: lines,
		flags: make([]uint8, len(lines)),
	}
	var s scanner
	prevInFence := false
	fenceStart := 0
	for i, line := range lines {
		lc := s.classify(line)
		c.flags[i] = lc.flags
		// Record fence spans from the scanner's open/close transitions. A line is
		// at most one transition: a closing-fence line drops inFence on the same
		// line that is still flagged fenced, and the next opener (even if
		// adjacent) is a separate open transition, so adjacent blocks stay
		// distinct.
		if s.inFence && !prevInFence {
			fenceStart = i
		} else if !s.inFence && prevInFence {
			c.fences = append(c.fences, FenceSpan{Start: fenceStart, End: i})
		}
		prevInFence = s.inFence
		// Store the sanitized text only when it actually differs from the
		// original. The fast path in sanitizeInline returns the original string
		// unchanged for ordinary lines, so this comparison is a cheap identity
		// check for the common case.
		if lc.sanitized != line {
			if c.sanitized == nil {
				c.sanitized = make(map[int]string)
			}
			c.sanitized[i] = lc.sanitized
		}
	}
	// A fence still open at EOF is unclosed; record it with End == -1.
	if s.inFence {
		c.fences = append(c.fences, FenceSpan{Start: fenceStart, End: -1})
	}
	return c
}

// lineClass is the per-line classification produced by the scanner: the packed
// context flags and the inline-sanitized text (which equals the original line
// unless an inline code span or comment was blanked).
type lineClass struct {
	flags     uint8
	sanitized string
}

// scanner holds the cross-line state needed to classify each line in context.
type scanner struct {
	inFence     bool
	fenceMarker string

	inHTMLBlock bool
	htmlType    int

	inComment   bool
	inParagraph bool
}

// classify advances the scanner by one line and returns that line's class.
func (s *scanner) classify(line string) lineClass {
	cols, firstIdx := indentColumns(line)
	isBlank := firstIdx == len(line)

	if lc, handled := s.continueOpenBlock(line, cols, isBlank); handled {
		return lc
	}
	return s.startLine(line, cols, isBlank)
}

// continueOpenBlock handles a line that lies inside a block opened on an earlier
// line (fenced code, HTML comment, or HTML block). It returns handled=false when
// no such block is open, or when a type 6/7 HTML block has just ended on this
// blank line and the line should be classified afresh by startLine.
func (s *scanner) continueOpenBlock(line string, cols int, isBlank bool) (lineClass, bool) {
	switch {
	// Fences take precedence over every other context; a comment or HTML start
	// inside a code block is literal.
	case s.inFence:
		if !isBlank && cols < 4 && isClosingFence(strings.TrimSpace(line), s.fenceMarker) {
			s.inFence = false
			s.fenceMarker = ""
		}
		s.inParagraph = false
		return lineClass{flags: flagFencedCode, sanitized: line}, true

	// Multi-line HTML comment: track until the closing "-->". The comment may
	// close mid-line and leave trailing prose; in that case the line is not a
	// pure comment line, so it is not flagged InHTMLComment and it reopens a
	// paragraph — mirroring startLine's handling of a fresh comment line.
	case s.inComment:
		sanitized, stillInComment, fullyComment := sanitizeInline(line, true)
		s.inComment = stillInComment
		lc := lineClass{sanitized: sanitized}
		if fullyComment {
			lc.flags = flagHTMLComment
			s.inParagraph = false
		} else {
			s.inParagraph = true
		}
		return lc, true

	// Open HTML block. Types 1 and 3–5 end on a delimiter line; types 6 and 7
	// end on a blank line, which is itself outside the block.
	case s.inHTMLBlock:
		if (s.htmlType == 6 || s.htmlType == 7) && isBlank {
			s.inHTMLBlock = false
			return lineClass{}, false
		}
		if s.htmlType >= 1 && s.htmlType <= 5 && htmlBlockEndsOnLine(line, s.htmlType) {
			s.inHTMLBlock = false
		}
		s.inParagraph = false
		return lineClass{flags: flagHTMLBlock, sanitized: line}, true
	}
	return lineClass{}, false
}

// startLine classifies a line that does not continue an already-open block.
func (s *scanner) startLine(line string, cols int, isBlank bool) lineClass {
	// Blank line outside any block closes an open paragraph.
	if isBlank {
		s.inParagraph = false
		return lineClass{sanitized: line}
	}

	// Indented code block: four or more columns of indentation, but only when it
	// does not continue an open paragraph (indented code cannot interrupt a
	// paragraph). Checked before openers because an indented line can open
	// neither a fence nor an HTML block.
	if cols >= 4 && !s.inParagraph {
		return lineClass{flags: flagIndentedCode, sanitized: line}
	}

	// Block openers are recognized only at an indentation below four columns.
	if cols < 4 {
		if lc, opened := s.tryOpenBlock(line); opened {
			return lc
		}
	}

	// Everything else is prose: paragraph text, headings, list items, lazy
	// paragraph continuations, and standalone HTML comments. Inline code spans
	// and inline comments are blanked in the sanitized text.
	sanitized, endedInComment, fullyComment := sanitizeInline(line, false)
	s.inComment = endedInComment
	lc := lineClass{sanitized: sanitized}
	if fullyComment {
		// The line was nothing but comment text — a standalone comment line.
		lc.flags = flagHTMLComment
		s.inParagraph = false
	} else {
		s.inParagraph = true
	}
	return lc
}

// tryOpenBlock attempts to open a fenced code block or an HTML block on line
// (which must be indented fewer than four columns). It returns opened=false when
// the line is not a block opener.
func (s *scanner) tryOpenBlock(line string) (lineClass, bool) {
	trimmed := strings.TrimSpace(line)

	if marker := openingFenceMarker(trimmed); marker != "" {
		s.inFence = true
		s.fenceMarker = marker
		s.inParagraph = false
		return lineClass{flags: flagFencedCode, sanitized: line}, true
	}

	if t := htmlBlockStart(trimmed, s.inParagraph); t != 0 {
		s.inHTMLBlock = true
		s.htmlType = t
		if t >= 1 && t <= 5 && htmlBlockEndsOnLine(trimmed, t) {
			s.inHTMLBlock = false
		}
		s.inParagraph = false
		return lineClass{flags: flagHTMLBlock, sanitized: line}, true
	}

	return lineClass{}, false
}
