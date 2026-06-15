package context

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/norlinga/loupe/internal/gitroot"
	"github.com/norlinga/loupe/internal/notes"
	"github.com/norlinga/loupe/internal/schema"
)

type Options struct {
	Now                   time.Time
	RecentlyModifiedSecs  int64
	RecentlyModifiedLimit int
	NoHidden              bool
}

func Enrich(node *schema.Node, opts Options) {
	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}
	if opts.RecentlyModifiedSecs <= 0 {
		opts.RecentlyModifiedSecs = 300
	}
	if opts.RecentlyModifiedLimit <= 0 {
		opts.RecentlyModifiedLimit = 20
	}
	root := gitroot.Nearest(node.Path)
	if root == "" {
		root = projectSearchStart(node.Path)
	}
	ctx := &schema.Context{
		ProjectType:          projectType(root),
		RecentlyModifiedSecs: opts.RecentlyModifiedSecs,
		RecentlyModified:     recentlyModified(root, opts.Now, time.Duration(opts.RecentlyModifiedSecs)*time.Second, opts.RecentlyModifiedLimit, opts.NoHidden),
	}
	if hasPath(filepath.Join(root, ".git")) {
		ctx.VCS = "git"
	}
	if agentNotes, ok := notes.ReadNearest(node.Path); ok {
		ctx.AgentNotes = agentNotes
	}
	node.Context = ctx
}

func projectSearchStart(path string) string {
	info, err := os.Lstat(path)
	if err == nil && !info.IsDir() {
		return filepath.Dir(path)
	}
	return path
}

func projectType(root string) string {
	switch {
	case hasPath(filepath.Join(root, "go.mod")):
		return "go"
	case hasPath(filepath.Join(root, "package.json")):
		return "javascript"
	case hasPath(filepath.Join(root, "Cargo.toml")):
		return "rust"
	case hasPath(filepath.Join(root, "pyproject.toml")):
		return "python"
	default:
		return ""
	}
}

type recentPath struct {
	path    string
	modTime time.Time
}

func recentlyModified(root string, now time.Time, window time.Duration, limit int, noHidden bool) []string {
	var paths []recentPath
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || path == root {
			return nil
		}
		name := entry.Name()
		if entry.IsDir() && (name == ".git" || name == ".loupe") {
			return filepath.SkipDir
		}
		if noHidden && strings.HasPrefix(name, ".") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil || info.IsDir() {
			return nil
		}
		if now.Sub(info.ModTime()) <= window {
			rel, err := filepath.Rel(root, path)
			if err == nil {
				paths = append(paths, recentPath{path: rel, modTime: info.ModTime()})
			}
		}
		return nil
	})
	sort.Slice(paths, func(i, j int) bool {
		if paths[i].modTime.Equal(paths[j].modTime) {
			return paths[i].path < paths[j].path
		}
		return paths[i].modTime.After(paths[j].modTime)
	})
	if limit > 0 && len(paths) > limit {
		paths = paths[:limit]
	}
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		result = append(result, path.path)
	}
	return result
}

func hasPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
