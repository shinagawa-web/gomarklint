package preprocess

import "strings"

type Context struct {
	lines     []string
	flags     []uint8
	sanitized map[int]string
	fences    []FenceSpan
}

// FenceSpan is the line range of one fenced code block.
// End is -1 when the fence is never closed (runs to end of file).
type FenceSpan struct {
	Start int
	End   int
}

const (
	flagFencedCode uint8 = 1 << iota
	flagIndentedCode
	flagHTMLBlock
	flagHTMLComment
)

func (c *Context) Len() int                  { return len(c.lines) }
func (c *Context) Line(i int) string         { return c.lines[i] }
func (c *Context) InFencedCode(i int) bool   { return c.flags[i]&flagFencedCode != 0 }
func (c *Context) InIndentedCode(i int) bool { return c.flags[i]&flagIndentedCode != 0 }
func (c *Context) InHTMLBlock(i int) bool    { return c.flags[i]&flagHTMLBlock != 0 }
func (c *Context) InHTMLComment(i int) bool  { return c.flags[i]&flagHTMLComment != 0 }
func (c *Context) FenceSpans() []FenceSpan   { return c.fences }

func (c *Context) Sanitized(i int) string {
	if s, ok := c.sanitized[i]; ok {
		return s
	}
	return c.lines[i]
}

// Scan classifies every line in a single pass. The input slice is borrowed and
// must not be mutated while the Context is in use.
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
		// Track fence open/close transitions so adjacent blocks stay distinct.
		// A closing-fence line is still flagged fenced, so the transition fires
		// on the line after the closer, not the closer itself.
		if s.inFence && !prevInFence {
			fenceStart = i
		} else if !s.inFence && prevInFence {
			c.fences = append(c.fences, FenceSpan{Start: fenceStart, End: i})
		}
		prevInFence = s.inFence
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

type lineClass struct {
	flags     uint8
	sanitized string
}

type scanner struct {
	inFence     bool
	fenceMarker string

	inHTMLBlock bool
	htmlType    int

	inComment   bool
	inParagraph bool
}

func (s *scanner) classify(line string) lineClass {
	cols, firstIdx := indentColumns(line)
	isBlank := firstIdx == len(line)

	if lc, handled := s.continueOpenBlock(line, cols, isBlank); handled {
		return lc
	}
	return s.startLine(line, cols, isBlank)
}

func (s *scanner) continueOpenBlock(line string, cols int, isBlank bool) (lineClass, bool) {
	switch {
	case s.inFence:
		if !isBlank && cols < 4 && isClosingFence(strings.TrimSpace(line), s.fenceMarker) {
			s.inFence = false
			s.fenceMarker = ""
		}
		s.inParagraph = false
		return lineClass{flags: flagFencedCode, sanitized: line}, true

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

func (s *scanner) startLine(line string, cols int, isBlank bool) lineClass {
	if isBlank {
		s.inParagraph = false
		return lineClass{sanitized: line}
	}

	// Indented code cannot interrupt a paragraph.
	if cols >= 4 && !s.inParagraph {
		return lineClass{flags: flagIndentedCode, sanitized: line}
	}

	if cols < 4 {
		if lc, opened := s.tryOpenBlock(line); opened {
			return lc
		}
	}

	sanitized, endedInComment, fullyComment := sanitizeInline(line, false)
	s.inComment = endedInComment
	lc := lineClass{sanitized: sanitized}
	if fullyComment {
		lc.flags = flagHTMLComment
		s.inParagraph = false
	} else {
		s.inParagraph = true
	}
	return lc
}

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
