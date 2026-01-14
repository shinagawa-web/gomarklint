package parser

import (
	"regexp"
	"strings"
)

var (
	// 1. [text](https://example.com)
	inlineLinkPattern = regexp.MustCompile(`\[[^\]]*\]\((https?://[^\s)]+)\)`)

	// 2. ![alt](https://example.com/image.png)
	imageLinkPattern = regexp.MustCompile(`!\[[^\]]*\]\((https?://[^\s)]+)\)`)

	// 3. https://example.com
	bareURLPattern = regexp.MustCompile(`(?m)^.*?(https?://[^\s<>()]+).*?$`)
)

type ExtractedLink struct {
	URL  string
	Line int
}

func ExtractExternalLinksWithLineNumbers(content string) []ExtractedLink {
	lines := strings.Split(content, "\n")
	patterns := []*regexp.Regexp{
		inlineLinkPattern,
		imageLinkPattern,
		bareURLPattern,
	}

	var results []ExtractedLink

	for i, line := range lines {
		seenInLine := make(map[string]bool) // Track URLs found in this line
		for _, re := range patterns {
			matches := re.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					url := match[1]
					// Only add if not already seen in this line
					if !seenInLine[url] {
						results = append(results, ExtractedLink{
							URL:  url,
							Line: i + 1, // 1-based line number
						})
						seenInLine[url] = true
					}
				}
			}
		}
	}
	return results
}
