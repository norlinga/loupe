package context

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/norlinga/loupe/internal/schema"
)

func TestEnrichDetectsProjectContextAndAgentNotes(t *testing.T) {
	root := t.TempDir()
	now := time.Unix(1_700_000_000, 0)
	mkdir(t, filepath.Join(root, ".git"))
	writeFile(t, filepath.Join(root, "go.mod"), "module example.test/project\n")
	writeFile(t, filepath.Join(root, "main.go"), "package main")
	writeFile(t, filepath.Join(root, "old.go"), "package main")
	writeFile(t, filepath.Join(root, ".loupe", "notes.json"), `{
  "schema_version": 1,
  "written_by": "test",
  "written_at": 1700000000,
  "notes": [
    {
      "kind": "gotcha",
      "summary": "Integration tests need fixtures",
      "paths": ["internal/fixtures"]
    }
  ]
}`)
	setModTime(t, filepath.Join(root, "go.mod"), now.Add(-10*time.Minute))
	setModTime(t, filepath.Join(root, "main.go"), now.Add(-60*time.Second))
	setModTime(t, filepath.Join(root, "old.go"), now.Add(-10*time.Minute))

	node := &schema.Node{Path: root}
	Enrich(node, Options{Now: now, RecentlyModifiedSecs: 300})

	if node.Context == nil {
		t.Fatal("context was not set")
	}
	if node.Context.VCS != "git" {
		t.Fatalf("VCS = %q, want git", node.Context.VCS)
	}
	if node.Context.ProjectType != "go" {
		t.Fatalf("ProjectType = %q, want go", node.Context.ProjectType)
	}
	if !reflect.DeepEqual(node.Context.RecentlyModified, []string{"main.go"}) {
		t.Fatalf("RecentlyModified = %#v, want main.go", node.Context.RecentlyModified)
	}
	if len(node.Context.AgentNotes) != 1 || node.Context.AgentNotes[0].Kind != "gotcha" {
		t.Fatalf("AgentNotes = %#v, want one gotcha", node.Context.AgentNotes)
	}
}

func TestEnrichSortsAndCapsRecentlyModified(t *testing.T) {
	root := t.TempDir()
	now := time.Unix(1_700_000_000, 0)
	mkdir(t, filepath.Join(root, ".git"))
	writeFile(t, filepath.Join(root, "go.mod"), "module example.test/project\n")
	writeFile(t, filepath.Join(root, "newest.go"), "package main")
	writeFile(t, filepath.Join(root, "a_tie.go"), "package main")
	writeFile(t, filepath.Join(root, "z_tie.go"), "package main")
	writeFile(t, filepath.Join(root, "old.go"), "package main")
	setModTime(t, filepath.Join(root, "go.mod"), now.Add(-30*time.Second))
	setModTime(t, filepath.Join(root, "newest.go"), now.Add(-10*time.Second))
	setModTime(t, filepath.Join(root, "a_tie.go"), now.Add(-20*time.Second))
	setModTime(t, filepath.Join(root, "z_tie.go"), now.Add(-20*time.Second))
	setModTime(t, filepath.Join(root, "old.go"), now.Add(-10*time.Minute))

	node := &schema.Node{Path: root}
	Enrich(node, Options{Now: now, RecentlyModifiedSecs: 300, RecentlyModifiedLimit: 3})

	want := []string{"newest.go", "a_tie.go", "z_tie.go"}
	if !reflect.DeepEqual(node.Context.RecentlyModified, want) {
		t.Fatalf("RecentlyModified = %#v, want %#v", node.Context.RecentlyModified, want)
	}
	if sort.StringsAreSorted([]string{node.Context.RecentlyModified[1], node.Context.RecentlyModified[2]}) == false {
		t.Fatalf("tie order was not lexical: %#v", node.Context.RecentlyModified)
	}
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
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

func setModTime(t *testing.T, path string, ts time.Time) {
	t.Helper()
	if err := os.Chtimes(path, ts, ts); err != nil {
		t.Fatal(err)
	}
}
