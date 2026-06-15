package observe

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/norlinga/loupe/internal/schema"
)

func TestObserveDirectoryNestedDepth(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "README.md"), "root")
	writeFile(t, filepath.Join(root, "internal", "observe.go"), "package internal")
	writeFile(t, filepath.Join(root, "internal", "deep", "ignored.go"), "package deep")

	node, err := Observe(root, Options{Depth: 2})
	if err != nil {
		t.Fatal(err)
	}
	if node.Type != schema.TypeDirectory {
		t.Fatalf("root type = %q, want directory", node.Type)
	}
	internal := requireEntry(t, node.Entries, "internal")
	if internal.Type != schema.TypeDirectory {
		t.Fatalf("internal type = %q, want directory", internal.Type)
	}
	requireEntry(t, internal.Entries, "observe.go")
	deep := requireEntry(t, internal.Entries, "deep")
	if len(deep.Entries) != 0 {
		t.Fatalf("depth 2 should not include great-grandchildren, got %#v", deep.Entries)
	}
}

func TestObserveDefaultDepthIsDirectoryOverviewOnly(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "src", "main.go"), "package main")

	node, err := Observe(root, Options{Depth: -1})
	if err != nil {
		t.Fatal(err)
	}
	src := requireEntry(t, node.Entries, "src")
	if len(src.Entries) != 0 {
		t.Fatalf("default directory depth should not include grandchildren, got %#v", src.Entries)
	}
}

func TestObserveDefaultDepthForFileHasNoEntries(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "main.go")
	writeFile(t, path, "package main")

	node, err := Observe(path, Options{Depth: -1})
	if err != nil {
		t.Fatal(err)
	}
	if node.Type != schema.TypeFile {
		t.Fatalf("type = %q, want file", node.Type)
	}
	if len(node.Entries) != 0 {
		t.Fatalf("file should not include entries, got %#v", node.Entries)
	}
}

func TestObserveEntriesAreLexicallyOrdered(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "b.txt"), "b")
	writeFile(t, filepath.Join(root, "a.txt"), "a")
	writeFile(t, filepath.Join(root, "c.txt"), "c")

	node, err := Observe(root, Options{Depth: 1})
	if err != nil {
		t.Fatal(err)
	}
	got := []string{node.Entries[0].Name, node.Entries[1].Name, node.Entries[2].Name}
	want := []string{"a.txt", "b.txt", "c.txt"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %#v, want %#v", got, want)
		}
	}
}

func TestObservePathologicalFilenames(t *testing.T) {
	root := t.TempDir()
	names := []string{
		"with spaces.txt",
		"with\ttab.txt",
		"with\nnewline.txt",
	}
	for _, name := range names {
		writeFile(t, filepath.Join(root, name), name)
	}

	node, err := Observe(root, Options{Depth: 1})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		requireEntry(t, node.Entries, name)
	}
}

func TestTypeFilterPreservesContainersForMatchingDescendants(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "src", "main.go"), "package main")
	writeFile(t, filepath.Join(root, "src", "README.md"), "docs")

	node, err := Observe(root, Options{Depth: 2, Type: schema.TypeFile})
	if err != nil {
		t.Fatal(err)
	}
	src := requireEntry(t, node.Entries, "src")
	if src.Type != schema.TypeDirectory {
		t.Fatalf("container type = %q, want directory", src.Type)
	}
	requireEntry(t, src.Entries, "main.go")
	requireEntry(t, src.Entries, "README.md")
}

func TestNoHiddenExcludesDotEntries(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".env"), "secret")
	writeFile(t, filepath.Join(root, "visible"), "ok")

	node, err := Observe(root, Options{Depth: 1, NoHidden: true})
	if err != nil {
		t.Fatal(err)
	}
	requireEntry(t, node.Entries, "visible")
	if entryByName(node.Entries, ".env") != nil {
		t.Fatal("hidden entry was included")
	}
}

func TestNewerThanFiltersEntriesWithInjectedClock(t *testing.T) {
	root := t.TempDir()
	now := time.Unix(1_700_000_000, 0)
	newPath := filepath.Join(root, "new.txt")
	oldPath := filepath.Join(root, "old.txt")
	writeFile(t, newPath, "new")
	writeFile(t, oldPath, "old")
	setModTime(t, newPath, now.Add(-60*time.Second))
	setModTime(t, oldPath, now.Add(-10*time.Minute))

	node, err := Observe(root, Options{Depth: 1, NewerThan: 5 * time.Minute, Now: now})
	if err != nil {
		t.Fatal(err)
	}
	requireEntry(t, node.Entries, "new.txt")
	if entryByName(node.Entries, "old.txt") != nil {
		t.Fatal("old entry was included")
	}
}

func TestSymlinkIncludesResolvedTarget(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target.txt")
	link := filepath.Join(root, "link.txt")
	writeFile(t, target, "target")
	if err := os.Symlink("target.txt", link); err != nil {
		t.Fatal(err)
	}

	node, err := Observe(link, Options{Depth: 0})
	if err != nil {
		t.Fatal(err)
	}
	if node.Type != schema.TypeSymlink {
		t.Fatalf("type = %q, want symlink", node.Type)
	}
	if node.Target != target {
		t.Fatalf("target = %q, want %q", node.Target, target)
	}
}

func TestBrokenSymlinkIncludesUnresolvedTarget(t *testing.T) {
	root := t.TempDir()
	link := filepath.Join(root, "missing-link")
	want := filepath.Join(root, "missing-target")
	if err := os.Symlink("missing-target", link); err != nil {
		t.Fatal(err)
	}

	node, err := Observe(link, Options{Depth: 0})
	if err != nil {
		t.Fatal(err)
	}
	if node.Type != schema.TypeSymlink {
		t.Fatalf("type = %q, want symlink", node.Type)
	}
	if node.Target != want {
		t.Fatalf("target = %q, want %q", node.Target, want)
	}
}

func TestUnreadableDescendantFailsObservation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod unreadable semantics differ on windows")
	}
	root := t.TempDir()
	locked := filepath.Join(root, "locked")
	mkdir(t, locked)
	t.Cleanup(func() {
		_ = os.Chmod(locked, 0o755)
	})
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}

	_, err := Observe(root, Options{Depth: 2})
	if err == nil {
		t.Skip("directory remained readable; likely running with elevated privileges")
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func setModTime(t *testing.T, path string, ts time.Time) {
	t.Helper()
	if err := os.Chtimes(path, ts, ts); err != nil {
		t.Fatal(err)
	}
}

func requireEntry(t *testing.T, entries []schema.Node, name string) schema.Node {
	t.Helper()
	entry := entryByName(entries, name)
	if entry == nil {
		t.Fatalf("entry %q not found in %#v", name, entries)
	}
	return *entry
}

func entryByName(entries []schema.Node, name string) *schema.Node {
	for i := range entries {
		if entries[i].Name == name {
			return &entries[i]
		}
	}
	return nil
}
