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
	seen := map[string]bool{}

	for i, line := range lines {
		for _, re := range patterns {
			matches := re.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					url := match[1]
					if !seen[url] {
						results = append(results, ExtractedLink{
							URL:  url,
							Line: i + 1, // 1-based line number
						})
						seen[url] = true
					}
				}
			}
		}
	}
	return results
}
