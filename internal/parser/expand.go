package parser

import (
	"github.com/shinagawa-web/gomarklint/internal/util"
	"os"
	"path/filepath"
	"strings"
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
			err := filepath.WalkDir(p, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return nil
				}

				if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
					return filepath.SkipDir
				}

				if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
					if util.ShouldIgnore(path, ignorePatterns) {
						return nil
					}
					results = append(results, path)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else if strings.HasSuffix(info.Name(), ".md") {
			if util.ShouldIgnore(p, ignorePatterns) {
				continue
			}
			results = append(results, p)
		}
	}

	return results, nil
}
