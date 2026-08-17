package preprocess

import "strings"

var htmlType1Names = []string{"script", "pre", "style", "textarea"}

var htmlType1CloseTags = []string{"</script>", "</style>", "</pre>", "</textarea>"}

var htmlBlockTags = map[string]bool{
	"address": true, "article": true, "aside": true, "base": true,
	"basefont": true, "blockquote": true, "body": true, "caption": true,
	"center": true, "col": true, "colgroup": true, "dd": true, "details": true,
	"dialog": true, "dir": true, "div": true, "dl": true, "dt": true,
	"fieldset": true, "figcaption": true, "figure": true, "footer": true,
	"form": true, "frame": true, "frameset": true, "h1": true, "h2": true,
	"h3": true, "h4": true, "h5": true, "h6": true, "head": true,
	"header": true, "hr": true, "html": true, "iframe": true, "legend": true,
	"li": true, "link": true, "main": true, "menu": true, "menuitem": true,
	"nav": true, "noframes": true, "ol": true, "optgroup": true, "option": true,
	"p": true, "param": true, "section": true, "summary": true, "table": true,
	"tbody": true, "td": true, "tfoot": true, "th": true, "thead": true,
	"title": true, "tr": true, "track": true, "ul": true,
}

func isASCIILetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isTagNameChar(c byte) bool {
	return isASCIILetter(c) || (c >= '0' && c <= '9') || c == '-'
}

// htmlBlockStart returns the HTML block type (1, 3, 4, 5, 6, or 7) opened by
// line, or 0 if it does not start an HTML block. Type 7 cannot interrupt a
// paragraph (inParagraph suppresses it).
func htmlBlockStart(line string, inParagraph bool) int {
	if len(line) < 2 || line[0] != '<' {
		return 0
	}
	if matchType1(line) {
		return 1
	}
	if strings.HasPrefix(line, "<?") {
		return 3
	}
	if strings.HasPrefix(line, "<![CDATA[") {
		return 5
	}
	if line[1] == '!' && len(line) >= 3 && isASCIILetter(line[2]) {
		return 4
	}
	if matchType6(line) {
		return 6
	}
	if !inParagraph && matchType7(line) {
		return 7
	}
	return 0
}

func matchType1(line string) bool {
	for _, name := range htmlType1Names {
		if len(line) < 1+len(name) {
			continue
		}
		if !strings.EqualFold(line[1:1+len(name)], name) {
			continue
		}
		rest := line[1+len(name):]
		if rest == "" {
			return true
		}
		switch rest[0] {
		case ' ', '\t', '>':
			return true
		}
	}
	return false
}

func matchType6(line string) bool {
	i := 1
	if i < len(line) && line[i] == '/' {
		i++
	}
	start := i
	if start >= len(line) || !isASCIILetter(line[start]) {
		return false
	}
	for i < len(line) && isTagNameChar(line[i]) {
		i++
	}
	name := strings.ToLower(line[start:i])
	if !htmlBlockTags[name] {
		return false
	}
	if i >= len(line) {
		return true
	}
	switch line[i] {
	case ' ', '\t', '>':
		return true
	case '/':
		return i+1 < len(line) && line[i+1] == '>'
	}
	return false
}

func matchType7(line string) bool {
	s := strings.TrimRight(line, " \t")
	if len(s) < 3 || s[0] != '<' || s[len(s)-1] != '>' {
		return false
	}
	inner := s[1 : len(s)-1]
	closing := false
	if strings.HasPrefix(inner, "/") {
		closing = true
		inner = inner[1:]
	}
	inner = strings.TrimSuffix(inner, "/")
	// A second tag delimiter means this is not a single complete tag.
	if strings.ContainsAny(inner, "<>") {
		return false
	}
	if inner == "" || !isASCIILetter(inner[0]) {
		return false
	}
	j := 0
	for j < len(inner) && isTagNameChar(inner[j]) {
		j++
	}
	name := strings.ToLower(inner[:j])
	for _, t := range htmlType1Names {
		if name == t {
			return false
		}
	}
	if closing {
		return strings.TrimSpace(inner[j:]) == ""
	}
	return true
}

func htmlBlockEndsOnLine(line string, blockType int) bool {
	switch blockType {
	case 1:
		lower := strings.ToLower(line)
		for _, close := range htmlType1CloseTags {
			if strings.Contains(lower, close) {
				return true
			}
		}
		return false
	case 3:
		return strings.Contains(line, "?>")
	case 4:
		return strings.Contains(line, ">")
	case 5:
		return strings.Contains(line, "]]>")
	}
	return false
}
