package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/norlinga/loupe/internal/schema"
)

func TestRunWritesJSONByDefault(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "main.go"), "package main")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := Run([]string{root}, &stdout, &stderr); err != nil {
		t.Fatalf("Run error = %v, stderr = %s", err, stderr.String())
	}
	var node schema.Node
	if err := json.Unmarshal(stdout.Bytes(), &node); err != nil {
		t.Fatal(err)
	}
	if node.Type != schema.TypeDirectory {
		t.Fatalf("type = %q, want directory", node.Type)
	}
	if len(node.Entries) != 1 || node.Entries[0].Name != "main.go" {
		t.Fatalf("entries = %#v", node.Entries)
	}
}

func TestRunFileOutputMatchesGolden(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "main.go")
	writeFile(t, path, "package main")
	setModTime(t, path, time.Unix(1_700_000_000, 0))
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := Run([]string{path}, &stdout, &stderr); err != nil {
		t.Fatalf("Run error = %v, stderr = %s", err, stderr.String())
	}
	assertGolden(t, "golden_file.json", stdout.String(), root)
}

func TestRunNestedOutputMatchesGolden(t *testing.T) {
	root := t.TempDir()
	readme := filepath.Join(root, "README.md")
	mainGo := filepath.Join(root, "src", "main.go")
	writeFile(t, readme, "readme")
	writeFile(t, mainGo, "package main")
	setModTime(t, readme, time.Unix(1_700_000_001, 0))
	setModTime(t, mainGo, time.Unix(1_700_000_002, 0))
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := Run([]string{root, "--depth", "2"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run error = %v, stderr = %s", err, stderr.String())
	}
	assertGolden(t, "golden_nested.json", stdout.String(), root)
}

func TestRunAcceptsFlagsAfterPath(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "src", "main.go"), "package main")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := Run([]string{root, "--depth", "2", "--type", "file"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run error = %v, stderr = %s", err, stderr.String())
	}
	var node schema.Node
	if err := json.Unmarshal(stdout.Bytes(), &node); err != nil {
		t.Fatal(err)
	}
	src := node.Entries[0]
	if src.Name != "src" || len(src.Entries) != 1 || src.Entries[0].Name != "main.go" {
		t.Fatalf("entries = %#v", node.Entries)
	}
}

func TestRunAcceptsDashPrefixedPathAfterDoubleDash(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "-dash")
	writeFile(t, path, "content")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := Run([]string{"--", path}, &stdout, &stderr); err != nil {
		t.Fatalf("Run error = %v, stderr = %s", err, stderr.String())
	}
	var node schema.Node
	if err := json.Unmarshal(stdout.Bytes(), &node); err != nil {
		t.Fatal(err)
	}
	if node.Name != "-dash" || node.Type != schema.TypeFile {
		t.Fatalf("node = %#v", node)
	}
}

func TestRunVersionDoesNotRequirePath(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := Run([]string{"--version"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run error = %v, stderr = %s", err, stderr.String())
	}
	if stdout.String() != "dev\n" {
		t.Fatalf("stdout = %q, want dev", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunRejectsInvalidType(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run([]string{".", "--type", "socket"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run returned nil error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !bytes.Contains(stderr.Bytes(), []byte("unsupported --type")) {
		t.Fatalf("stderr = %q, want unsupported type message", stderr.String())
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

func assertGolden(t *testing.T, name string, got string, root string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	want := string(data)
	want = strings.ReplaceAll(want, "<ROOT>", filepath.ToSlash(root))
	want = strings.ReplaceAll(want, "<ROOT_NAME>", filepath.Base(root))
	want = replaceVolatileDirectoryFields(want, got)
	if got != want {
		t.Fatalf("golden mismatch (-want +got)\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func replaceVolatileDirectoryFields(want string, got string) string {
	var gotNode schema.Node
	if err := json.Unmarshal([]byte(got), &gotNode); err != nil {
		return want
	}
	var replacements []struct {
		size     int64
		modified int64
	}
	collectDirectoryFields(gotNode, &replacements)
	for _, replacement := range replacements {
		want = strings.Replace(want, "\"<DIR_SIZE>\"", fmt.Sprintf("%d", replacement.size), 1)
		want = strings.Replace(want, "\"<DIR_MODIFIED>\"", fmt.Sprintf("%d", replacement.modified), 1)
	}
	return want
}

func collectDirectoryFields(node schema.Node, fields *[]struct {
	size     int64
	modified int64
}) {
	if node.Type == schema.TypeDirectory {
		*fields = append(*fields, struct {
			size     int64
			modified int64
		}{size: node.SizeBytes, modified: node.ModifiedUnix})
	}
	for _, entry := range node.Entries {
		collectDirectoryFields(entry, fields)
	}
}
