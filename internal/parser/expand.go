package parser

import (
	"os"
	"path/filepath"
	"strings"
)

func ExpandPaths(paths []string) ([]string, error) {
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
					results = append(results, path)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else if strings.HasSuffix(info.Name(), ".md") {
			results = append(results, p)
		}
	}

	return results, nil
}
