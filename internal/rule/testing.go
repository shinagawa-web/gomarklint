package rule

import (
	"path/filepath"
	"runtime"
)

// getTestFilePath returns the absolute path to a test file relative to the project root.
func getTestFilePath(rel string) string {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "../..")
	return filepath.Join(projectRoot, rel)
}
