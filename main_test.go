package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestLoupeBinaryServesMCP(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "main.go")
	writeFile(t, path, "package main")
	exe := filepath.Join(t.TempDir(), "loupe")

	build := exec.Command("go", "build", "-o", exe, ".")
	build.Dir = "."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}

	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"loupe_observe","arguments":{"path":"` + filepath.ToSlash(path) + `"}}}` + "\n"
	run := exec.Command(exe, "--mcp")
	run.Stdin = strings.NewReader(input)
	var stderr bytes.Buffer
	run.Stderr = &stderr
	out, err := run.Output()
	if err != nil {
		t.Fatalf("loupe --mcp failed: %v\nstderr: %s", err, stderr.String())
	}
	if !bytes.Contains(out, []byte("loupe_observe")) && !bytes.Contains(out, []byte("main.go")) {
		t.Fatalf("unexpected MCP output: %s", out)
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
