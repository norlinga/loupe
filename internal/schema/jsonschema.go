package schema

const JSONSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://github.com/norlinga/loupe/docs/loupe.schema.json",
  "title": "loupe output",
  "type": "object",
  "required": [
    "path",
    "name",
    "type",
    "size_bytes",
    "modified_unix",
    "permissions"
  ],
  "additionalProperties": false,
  "properties": {
    "path": {
      "type": "string",
      "description": "Absolute filesystem path for this node."
    },
    "name": {
      "type": "string",
      "description": "Base filename for this node."
    },
    "type": {
      "type": "string",
      "enum": ["file", "directory", "symlink", "other"]
    },
    "size_bytes": {
      "type": "integer",
      "minimum": 0
    },
    "modified_unix": {
      "type": "integer",
      "description": "Unix epoch timestamp in seconds."
    },
    "permissions": {
      "type": "string",
      "pattern": "^[0-7]{3,4}$",
      "description": "Plain octal permission bits."
    },
    "entry_count": {
      "type": "integer",
      "minimum": 0
    },
    "extension": {
      "type": "string"
    },
    "target": {
      "type": "string",
      "description": "Resolved symlink target path, or unresolved absolute target for broken symlinks."
    },
    "entries": {
      "type": "array",
      "items": {
        "$ref": "#"
      }
    },
    "context": {
      "$ref": "#/$defs/context"
    }
  },
  "$defs": {
    "context": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "vcs": {
          "type": "string"
        },
        "project_type": {
          "type": "string"
        },
        "recently_modified_secs": {
          "type": "integer",
          "minimum": 0
        },
        "recently_modified": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "agent_notes": {
          "type": "array",
          "items": {
            "$ref": "#/$defs/agent_note"
          }
        }
      }
    },
    "agent_note": {
      "type": "object",
      "required": ["kind", "summary"],
      "additionalProperties": false,
      "properties": {
        "kind": {
          "type": "string",
          "enum": ["fix", "warning", "decision", "gotcha", "observation"]
        },
        "summary": {
          "type": "string"
        },
        "paths": {
          "type": "array",
          "items": {
            "type": "string"
          }
        }
      }
    }
  }
}`
