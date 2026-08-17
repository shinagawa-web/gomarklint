package preprocess

// indentColumns returns the indentation of line in columns (tabs expand to the
// next multiple of 4) and the byte index of the first non-whitespace character.
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

// Known limitation: indentation is measured from the start of the line, not
// relative to an enclosing list item or block quote. Lines inside a list that
// happen to be indented ≥4 columns may be mis-flagged as InIndentedCode.
// Downstream rules should not treat InIndentedCode as authoritative inside
// list/blockquote contexts.
