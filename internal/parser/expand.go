package parser

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/shinagawa-web/gomarklint/internal/util"
)

// ExpandPaths takes a list of file or directory paths and returns a slice of
// all Markdown (.md) file paths found.
//
// It handles the following behavior:
//   - Recursively searches directories for files ending in .md
//   - Skips hidden directories (e.g., .git/)
//   - Ignores non-existent paths without failing
//   - Does not follow symbolic links
//
// Example:
//
//	input:  []string{"docs", "README.md"}
//	output: ["docs/a.md", "docs/sub/b.md", "README.md"]
func ExpandPaths(paths []string, ignorePatterns []string) ([]string, error) {
	var results []string

	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}

		if info.IsDir() {
			dirResults, err := expandDirectory(p, ignorePatterns)
			if err != nil {
				return nil, err
			}
			results = append(results, dirResults...)
		} else if strings.HasSuffix(info.Name(), ".md") {
			if !util.ShouldIgnore(p, ignorePatterns) {
				results = append(results, p)
			}
		}
	}

	return results, nil
}

func expandDirectory(root string, ignorePatterns []string) ([]string, error) {
	var results []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			if shouldSkipDirectory(path, root, d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if isMarkdownFile(d.Name()) && !util.ShouldIgnore(path, ignorePatterns) {
			results = append(results, path)
		}
		return nil
	})

	return results, err
}

func shouldSkipDirectory(path, root, name string) bool {
	return path != root && strings.HasPrefix(name, ".")
}

func isMarkdownFile(name string) bool {
	return strings.HasSuffix(name, ".md")
}
