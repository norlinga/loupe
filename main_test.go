package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/norlinga/loupe/internal/schema"
)

func TestLoupeBinaryWritesJSON(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "main.go"), "package main")
	exe := filepath.Join(t.TempDir(), "loupe")

	build := exec.Command("go", "build", "-o", exe, ".")
	build.Dir = "."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}

	run := exec.Command(exe, root, "--depth", "1")
	out, err := run.Output()
	if err != nil {
		t.Fatalf("loupe failed: %v", err)
	}
	var node schema.Node
	if err := json.Unmarshal(out, &node); err != nil {
		t.Fatal(err)
	}
	if node.Type != schema.TypeDirectory || len(node.Entries) != 1 || node.Entries[0].Name != "main.go" {
		t.Fatalf("node = %#v", node)
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
