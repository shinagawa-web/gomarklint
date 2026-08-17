package file

import (
	"os"
	"strings"
)

func ReadFile(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func StripFrontmatter(content string) (string, int) {
	lines := strings.Split(content, "\n")
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				skip := i + 1
				for skip < len(lines) && strings.TrimSpace(lines[skip]) == "" {
					skip++
				}
				return strings.Join(lines[skip:], "\n"), skip
			}
		}
	}
	return content, 0
}
