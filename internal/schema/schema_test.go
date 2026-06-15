package schema

import (
	"encoding/json"
	"testing"
)

func TestNodeJSONUsesTypedFieldsAndOmitsEmptyOptionalFields(t *testing.T) {
	node := Node{
		Path:         "/tmp/loupe/main.go",
		Name:         "main.go",
		Type:         TypeFile,
		SizeBytes:    42,
		ModifiedUnix: 1_700_000_000,
		Permissions:  "644",
		Extension:    ".go",
	}

	data, err := json.Marshal(node)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, ok := decoded["size_bytes"].(float64); !ok {
		t.Fatalf("size_bytes = %#v, want JSON number", decoded["size_bytes"])
	}
	if _, ok := decoded["modified_unix"].(float64); !ok {
		t.Fatalf("modified_unix = %#v, want JSON number", decoded["modified_unix"])
	}
	if decoded["permissions"] != "644" {
		t.Fatalf("permissions = %#v, want octal string", decoded["permissions"])
	}
	if _, ok := decoded["entries"]; ok {
		t.Fatal("entries should be omitted for file nodes with no entries")
	}
	if _, ok := decoded["context"]; ok {
		t.Fatal("context should be omitted when unset")
	}
}
