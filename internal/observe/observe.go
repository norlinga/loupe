package observe

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/norlinga/loupe/internal/schema"
)

type Options struct {
	Depth     int
	Type      schema.EntryType
	NewerThan time.Duration
	NoHidden  bool
	Now       time.Time
}

func Observe(path string, opts Options) (*schema.Node, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return nil, err
	}
	if opts.Depth < 0 {
		if info.IsDir() {
			opts.Depth = 1
		} else {
			opts.Depth = 0
		}
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}
	node, _, err := observePath(abs, info, opts, opts.Depth, true)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func ParseType(value string) (schema.EntryType, error) {
	switch strings.ToLower(value) {
	case "":
		return "", nil
	case "file":
		return schema.TypeFile, nil
	case "dir", "directory":
		return schema.TypeDirectory, nil
	case "symlink":
		return schema.TypeSymlink, nil
	default:
		return "", fmt.Errorf("unsupported type %q", value)
	}
}

func observePath(path string, info os.FileInfo, opts Options, depth int, root bool) (*schema.Node, bool, error) {
	node := nodeFromInfo(path, info)
	if node.Type == schema.TypeSymlink {
		node.Target = symlinkTarget(path)
	}
	if node.Type == schema.TypeDirectory && depth > 0 {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, false, err
		}
		for _, entry := range entries {
			if opts.NoHidden && strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			childPath := filepath.Join(path, entry.Name())
			childInfo, err := os.Lstat(childPath)
			if err != nil {
				return nil, false, err
			}
			child, include, err := observePath(childPath, childInfo, opts, depth-1, false)
			if err != nil {
				return nil, false, err
			}
			if include {
				node.Entries = append(node.Entries, *child)
			}
		}
		node.EntryCount = len(node.Entries)
	}
	return node, root || shouldInclude(node, opts), nil
}

func nodeFromInfo(path string, info os.FileInfo) *schema.Node {
	node := &schema.Node{
		Path:         path,
		Name:         filepath.Base(path),
		Type:         typeFromMode(info.Mode()),
		SizeBytes:    info.Size(),
		ModifiedUnix: info.ModTime().Unix(),
		Permissions:  fmt.Sprintf("%03o", info.Mode().Perm()),
	}
	if node.Type == schema.TypeFile {
		node.Extension = filepath.Ext(info.Name())
	}
	return node
}

func shouldInclude(node *schema.Node, opts Options) bool {
	typeMatches := opts.Type == "" || node.Type == opts.Type
	timeMatches := opts.NewerThan <= 0 || opts.Now.Sub(nodeTime(node)) <= opts.NewerThan
	if node.Type == schema.TypeDirectory && len(node.Entries) > 0 {
		return true
	}
	return typeMatches && timeMatches
}

func nodeTime(node *schema.Node) time.Time {
	return time.Unix(node.ModifiedUnix, 0)
}

func typeFromMode(mode os.FileMode) schema.EntryType {
	switch {
	case mode&os.ModeSymlink != 0:
		return schema.TypeSymlink
	case mode.IsRegular():
		return schema.TypeFile
	case mode.IsDir():
		return schema.TypeDirectory
	default:
		return schema.TypeOther
	}
}

func symlinkTarget(path string) string {
	target, err := os.Readlink(path)
	if err != nil {
		return ""
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}
	resolved, err := filepath.EvalSymlinks(target)
	if err == nil {
		return resolved
	}
	return filepath.Clean(target)
}
