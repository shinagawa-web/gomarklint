package file

import (
	"github.com/bmatcuk/doublestar/v4"
)

// shouldIgnore checks if the given path matches any of the ignore patterns.
func ShouldIgnore(path string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := doublestar.PathMatch(pattern, path)
		if err == nil && matched {
			return true
		}
	}
	return false
}
