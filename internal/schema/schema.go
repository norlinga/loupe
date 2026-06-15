package schema

type EntryType string

const (
	TypeFile      EntryType = "file"
	TypeDirectory EntryType = "directory"
	TypeSymlink   EntryType = "symlink"
	TypeOther     EntryType = "other"
)

type Node struct {
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	Type         EntryType `json:"type"`
	SizeBytes    int64     `json:"size_bytes"`
	ModifiedUnix int64     `json:"modified_unix"`
	Permissions  string    `json:"permissions"`
	EntryCount   int       `json:"entry_count,omitempty"`
	Extension    string    `json:"extension,omitempty"`
	Target       string    `json:"target,omitempty"`
	Entries      []Node    `json:"entries,omitempty"`
	Context      *Context  `json:"context,omitempty"`
}

type Context struct {
	VCS                  string      `json:"vcs,omitempty"`
	ProjectType          string      `json:"project_type,omitempty"`
	RecentlyModifiedSecs int64       `json:"recently_modified_secs,omitempty"`
	RecentlyModified     []string    `json:"recently_modified,omitempty"`
	AgentNotes           []AgentNote `json:"agent_notes,omitempty"`
}

type AgentNote struct {
	Kind    string   `json:"kind"`
	Summary string   `json:"summary"`
	Paths   []string `json:"paths,omitempty"`
}
