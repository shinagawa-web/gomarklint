package testutil

import (
	"path/filepath"
	"runtime"
)

// getTestFilePath resolves relative path from project root
func GetTestFilePath(rel string) string {
	_, filename, _, _ := runtime.Caller(1)
	projectRoot := filepath.Join(filepath.Dir(filename), "../..")
	return filepath.Join(projectRoot, rel)
}
