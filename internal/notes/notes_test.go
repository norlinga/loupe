package notes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadNearestReturnsProjectRootNotes(t *testing.T) {
	root := t.TempDir()
	mkdir(t, filepath.Join(root, ".git"))
	writeFile(t, filepath.Join(root, ".loupe", "notes.json"), `{
  "schema_version": 1,
  "written_by": "test",
  "written_at": 1700000000,
  "notes": [{"kind":"decision","summary":"Use nested output"}]
}`)
	nested := filepath.Join(root, "internal", "observe")
	mkdir(t, nested)

	notes, ok := ReadNearest(nested)
	if !ok {
		t.Fatal("ReadNearest returned ok=false")
	}
	if len(notes) != 1 || notes[0].Summary != "Use nested output" {
		t.Fatalf("notes = %#v", notes)
	}
}

func TestReadNearestSkipsMalformedNotes(t *testing.T) {
	root := t.TempDir()
	mkdir(t, filepath.Join(root, ".git"))
	writeFile(t, filepath.Join(root, ".loupe", "notes.json"), `{not json`)

	notes, ok := ReadNearest(root)
	if ok {
		t.Fatalf("ok = true, notes = %#v", notes)
	}
}

func TestEmbeddedJSONSchemaMatchesDocs(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "docs", "notes.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(JSONSchema) != strings.TrimSpace(string(data)) {
		t.Fatal("embedded JSONSchema differs from docs/notes.schema.json")
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
