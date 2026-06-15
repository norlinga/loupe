package gitroot

import (
	"os"
	"path/filepath"
)

// Nearest walks up from path to find the nearest directory containing .git.
// If path is a file, the walk starts from its parent. Returns "" if none found.
func Nearest(path string) string {
	dir := path
	info, err := os.Lstat(path)
	if err == nil && !info.IsDir() {
		dir = filepath.Dir(path)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
