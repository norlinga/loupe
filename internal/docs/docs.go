package docs

import (
	"github.com/norlinga/loupe/internal/notes"
	"github.com/norlinga/loupe/internal/schema"
)

func OutputSchema() string {
	return schema.JSONSchema
}

func NotesSchema() string {
	return notes.JSONSchema
}
