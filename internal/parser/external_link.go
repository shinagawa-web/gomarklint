package parser

import (
	"regexp"
)

var (
	// 1. [text](https://example.com)
	inlineLinkPattern = regexp.MustCompile(`\[[^\]]*\]\((https?://[^\s)]+)\)`)

	// 2. ![alt](https://example.com/image.png)
	imageLinkPattern = regexp.MustCompile(`!\[[^\]]*\]\((https?://[^\s)]+)\)`)

	// 3. https://example.com
	bareURLPattern = regexp.MustCompile(`(?m)^.*?(https?://[^\s<>()]+).*?$`)
)

func ExtractExternalLinks(text string) []string {
	links := make(map[string]bool)

	for _, re := range []*regexp.Regexp{inlineLinkPattern, imageLinkPattern, bareURLPattern} {
		matches := re.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				links[match[1]] = true
			}
		}
	}

	result := make([]string, 0, len(links))
	for url := range links {
		result = append(result, url)
	}
	return result
}
