package file

import (
	"os"
	"path/filepath"
	"strings"
)

func ExpandPaths(paths []string, ignorePatterns []string) []string {
	var results []string

	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}

		if info.IsDir() {
			results = append(results, expandDirectory(p, ignorePatterns)...)
		} else if strings.HasSuffix(info.Name(), ".md") {
			if !ShouldIgnore(p, ignorePatterns) {
				results = append(results, p)
			}
		}
	}

	return results
}

func expandDirectory(root string, ignorePatterns []string) []string {
	var results []string

	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			if shouldSkipDirectory(path, root, d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if isMarkdownFile(d.Name()) && !ShouldIgnore(path, ignorePatterns) {
			results = append(results, path)
		}
		return nil
	})

	return results
}

func shouldSkipDirectory(path, root, name string) bool {
	return path != root && strings.HasPrefix(name, ".")
}

func isMarkdownFile(name string) bool {
	return strings.HasSuffix(name, ".md")
}
