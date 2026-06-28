package preprocess

// indentColumns returns the indentation of line measured in columns and the
// byte index of the first non-whitespace character. A tab advances to the next
// multiple of four columns, matching CommonMark's tab-stop handling. If the
// line is empty or contains only whitespace, firstIdx equals len(line) and the
// line should be treated as blank.
//
// Only spaces and tabs count as indentation here; a leading carriage return is
// not expected because callers split on newlines.
func indentColumns(line string) (cols, firstIdx int) {
	col := 0
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case ' ':
			col++
		case '\t':
			col += 4 - (col % 4)
		default:
			return col, i
		}
	}
	return col, len(line)
}

// Indented code blocks (CommonMark §4.4) are detected in Scan rather than here,
// because the rule is stateful: a line indented four or more columns is code
// only when it does not continue an open paragraph (an indented code block
// cannot interrupt a paragraph). The paragraph state is tracked across lines in
// Scan.
//
// Known limitation (Phase 1): indentation is measured from the start of the
// line, not relative to an enclosing list item or block quote. Inside a list,
// continuation text is commonly indented several columns and is list content,
// not an indented code block. Scan does not model container blocks, so such
// lines may be mis-flagged as InIndentedCode. Downstream rules should not treat
// InIndentedCode as authoritative inside list/blockquote contexts.
