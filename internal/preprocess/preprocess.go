// Package preprocess provides a single shared pre-pass over a Markdown file's
// lines, classifying each line by the context it lives in (fenced code,
// indented code, HTML block, HTML comment) and producing a sanitized copy with
// inline code spans and inline HTML comments blanked out.
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
package preprocess

import "strings"

// LineContext records, for one source line, which Markdown contexts it lives in
// and a sanitized copy of its text.
//
// The boolean flags describe structural block membership and are mutually
// exclusive: a line is in at most one of fenced code, indented code, an HTML
// block, or an HTML comment. They are exposed individually rather than via a
// single "skippable" convenience flag because different rules skip different
// subsets — for example, rules implementing the markdownlint divergences in
// issue #337 (max-line-length, no-hard-tabs) deliberately scan inside fenced
// code blocks.
//
// Sanitized is an inline-level transformation only: inline code spans and inline
// HTML comments are replaced by spaces (length-preserving), so a rule can scan
// Sanitized without matching content that lives in those spans. Block-level code
// (fenced and indented) is NOT blanked in Sanitized — it is left verbatim so
// that rules which want to inspect code-block contents still can; use the flags
// to skip block code instead. For lines flagged InHTMLComment, Sanitized is
// fully blank because the entire line was comment text.
type LineContext struct {
	// Original is the line exactly as it appeared in the input.
	Original string
	// Sanitized is Original with inline code spans and inline HTML comments
	// replaced by spaces. For block-code and HTML-block lines it equals
	// Original; for fully-commented lines it is all whitespace.
	Sanitized string

	// InFencedCode is true for every line of a fenced code block, including the
	// opening and closing fence lines. If a fence is never closed, every line
	// from the opener to end of file is flagged, which prevents downstream rules
	// from mispairing fences inside the unclosed region.
	InFencedCode bool
	// InIndentedCode is true for lines inside an indented code block
	// (CommonMark §4.4). See the limitation noted in indented_code.go.
	InIndentedCode bool
	// InHTMLBlock is true for lines inside an HTML block of CommonMark types 1
	// and 3–7 (e.g. <div>…</div>). HTML comments are reported via
	// InHTMLComment, not this flag.
	InHTMLBlock bool
	// InHTMLComment is true for lines whose entire content is an HTML comment —
	// a standalone <!-- … --> line or a continuation line of a multi-line
	// comment. A prose line with a trailing inline comment is not flagged here;
	// its comment is blanked in Sanitized instead.
	InHTMLComment bool
}

// Scan classifies every line of a Markdown file in a single pass and returns one
// LineContext per input line, in order. The input is expected to have already
// had YAML front matter stripped by the caller (the linter does this centrally),
// so Scan does not handle front matter.
func Scan(lines []string) []LineContext {
	result := make([]LineContext, len(lines))
	var s scanner
	for i, line := range lines {
		result[i] = s.classify(line)
	}
	return result
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

// classify advances the scanner by one line and returns that line's context.
func (s *scanner) classify(line string) LineContext {
	cols, firstIdx := indentColumns(line)
	isBlank := firstIdx == len(line)

	if ctx, handled := s.continueOpenBlock(line, cols, isBlank); handled {
		return ctx
	}
	return s.startLine(line, cols, isBlank)
}

// continueOpenBlock handles a line that lies inside a block opened on an earlier
// line (fenced code, HTML comment, or HTML block). It returns handled=false when
// no such block is open, or when a type 6/7 HTML block has just ended on this
// blank line and the line should be classified afresh by startLine.
func (s *scanner) continueOpenBlock(line string, cols int, isBlank bool) (LineContext, bool) {
	switch {
	// Fences take precedence over every other context; a comment or HTML start
	// inside a code block is literal.
	case s.inFence:
		if !isBlank && cols < 4 && isClosingFence(strings.TrimSpace(line), s.fenceMarker) {
			s.inFence = false
			s.fenceMarker = ""
		}
		s.inParagraph = false
		return LineContext{Original: line, Sanitized: line, InFencedCode: true}, true

	// Multi-line HTML comment: track until the closing "-->". The comment may
	// close mid-line and leave trailing prose; in that case the line is not a
	// pure comment line, so it is not flagged InHTMLComment and it reopens a
	// paragraph — mirroring startLine's handling of a fresh comment line.
	case s.inComment:
		sanitized, stillInComment, fullyComment := sanitizeInline(line, true)
		s.inComment = stillInComment
		ctx := LineContext{Original: line, Sanitized: sanitized}
		if fullyComment {
			ctx.InHTMLComment = true
			s.inParagraph = false
		} else {
			s.inParagraph = true
		}
		return ctx, true

	// Open HTML block. Types 1 and 3–5 end on a delimiter line; types 6 and 7
	// end on a blank line, which is itself outside the block.
	case s.inHTMLBlock:
		if (s.htmlType == 6 || s.htmlType == 7) && isBlank {
			s.inHTMLBlock = false
			return LineContext{}, false
		}
		if s.htmlType >= 1 && s.htmlType <= 5 && htmlBlockEndsOnLine(line, s.htmlType) {
			s.inHTMLBlock = false
		}
		s.inParagraph = false
		return LineContext{Original: line, Sanitized: line, InHTMLBlock: true}, true
	}
	return LineContext{}, false
}

// startLine classifies a line that does not continue an already-open block.
func (s *scanner) startLine(line string, cols int, isBlank bool) LineContext {
	// Blank line outside any block closes an open paragraph.
	if isBlank {
		s.inParagraph = false
		return LineContext{Original: line, Sanitized: line}
	}

	// Indented code block: four or more columns of indentation, but only when it
	// does not continue an open paragraph (indented code cannot interrupt a
	// paragraph). Checked before openers because an indented line can open
	// neither a fence nor an HTML block.
	if cols >= 4 && !s.inParagraph {
		return LineContext{Original: line, Sanitized: line, InIndentedCode: true}
	}

	// Block openers are recognized only at an indentation below four columns.
	if cols < 4 {
		if ctx, opened := s.tryOpenBlock(line); opened {
			return ctx
		}
	}

	// Everything else is prose: paragraph text, headings, list items, lazy
	// paragraph continuations, and standalone HTML comments. Inline code spans
	// and inline comments are blanked in Sanitized.
	sanitized, endedInComment, fullyComment := sanitizeInline(line, false)
	s.inComment = endedInComment
	ctx := LineContext{Original: line, Sanitized: sanitized}
	if fullyComment {
		// The line was nothing but comment text — a standalone comment line.
		ctx.InHTMLComment = true
		s.inParagraph = false
	} else {
		s.inParagraph = true
	}
	return ctx
}

// tryOpenBlock attempts to open a fenced code block or an HTML block on line
// (which must be indented fewer than four columns). It returns opened=false when
// the line is not a block opener.
func (s *scanner) tryOpenBlock(line string) (LineContext, bool) {
	trimmed := strings.TrimSpace(line)

	if marker := openingFenceMarker(trimmed); marker != "" {
		s.inFence = true
		s.fenceMarker = marker
		s.inParagraph = false
		return LineContext{Original: line, Sanitized: line, InFencedCode: true}, true
	}

	if t := htmlBlockStart(trimmed, s.inParagraph); t != 0 {
		s.inHTMLBlock = true
		s.htmlType = t
		if t >= 1 && t <= 5 && htmlBlockEndsOnLine(trimmed, t) {
			s.inHTMLBlock = false
		}
		s.inParagraph = false
		return LineContext{Original: line, Sanitized: line, InHTMLBlock: true}, true
	}

	return LineContext{}, false
}
