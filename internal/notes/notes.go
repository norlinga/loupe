package notes

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/norlinga/loupe/internal/gitroot"
	"github.com/norlinga/loupe/internal/schema"
)

type file struct {
	SchemaVersion int                `json:"schema_version"`
	WrittenBy     string             `json:"written_by"`
	WrittenAt     int64              `json:"written_at"`
	Notes         []schema.AgentNote `json:"notes"`
}

func ReadNearest(path string) ([]schema.AgentNote, bool) {
	root := gitroot.Nearest(path)
	if root == "" {
		return nil, false
	}
	data, err := os.ReadFile(filepath.Join(root, ".loupe", "notes.json"))
	if err != nil {
		return nil, false
	}
	var parsed file
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, false
	}
	if len(parsed.Notes) == 0 {
		return nil, false
	}
	return parsed.Notes, true
}

