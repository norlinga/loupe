# loupe

`loupe` observes local filesystem paths and emits typed JSON for agent workflows. It is intended to replace agent use of `ls`, `find`, and `stat` where the agent needs structured filesystem state rather than human-formatted text.

```sh
loupe <path> [flags]
```

JSON is the default output. Pipe it to `jq` for querying:

```sh
loupe . --depth 2 | jq '.entries[] | select(.type == "file")'
loupe ./internal --context | jq '.context'
loupe ./main.go
```

## Harness Quickstart

Build and verify locally:

```sh
make build VERSION=dev
./bin/loupe --help
./bin/loupe --schema
./bin/loupe --notes-schema
```

Use `loupe --mcp` for stdio MCP integration. See [docs/harness.md](docs/harness.md) for a full copy-paste setup, including MCP configuration and the reusable Codex skill.

## Flags

- `--depth N`: recurse `N` levels deep. Defaults to `1` for directories and `0` for files.
- `--type file|dir|directory|symlink`: filter emitted entries by type.
- `--newer-than N`: emit entries modified in the last `N` seconds.
- `--no-hidden`: exclude dotfile entries.
- `--context`: include project context such as VCS, project type, recently modified files, and agent notes.
- `--human`: print a minimal human-readable tree.
- `--schema`: print the JSON Schema and exit.
- `--notes-schema`: print the `.loupe/notes.json` JSON Schema and exit.
- `--version`: print the build version and exit.
- `--mcp`: serve loupe as a stdio MCP server.

Flags may appear before or after the path. Use `--` before paths that begin with `-`.

## Output

Directory recursion is nested. Each directory entry may contain its own `entries` array up to the requested depth.

```json
{
  "path": "/repo",
  "name": "repo",
  "type": "directory",
  "size_bytes": 4096,
  "modified_unix": 1718123456,
  "permissions": "755",
  "entry_count": 1,
  "entries": [
    {
      "path": "/repo/main.go",
      "name": "main.go",
      "type": "file",
      "size_bytes": 2048,
      "modified_unix": 1718120000,
      "permissions": "644",
      "extension": ".go"
    }
  ]
}
```

Entries are emitted in lexical filename order. If traversal encounters an unreadable descendant, `loupe` returns an error instead of partial JSON. Broken symlinks are still reported as symlinks with their unresolved target path.

See [docs/schema.md](docs/schema.md) and [docs/loupe.schema.json](docs/loupe.schema.json) for the full output contract.

Distributed binaries also expose the schema directly:

```sh
loupe --schema
```

## Agent Notes

When `--context` is used, `loupe` looks for `.loupe/notes.json` at the nearest git project root. Malformed notes are skipped silently.

```json
{
  "schema_version": 1,
  "written_by": "agent-session-id",
  "written_at": 1718123456,
  "notes": [
    {
      "kind": "gotcha",
      "summary": "Integration tests require a local Postgres instance",
      "paths": ["internal/db"]
    }
  ]
}
```

Distributed binaries expose this authoring contract directly:

```sh
loupe --notes-schema
```

## MCP

`loupe --mcp` starts a stdio MCP server exposing these tools:

- `loupe_observe`: observes a local filesystem path and returns loupe JSON as text content.
- `loupe_output_schema`: returns the output JSON Schema.
- `loupe_notes_schema`: returns the `.loupe/notes.json` JSON Schema.

Tool arguments:

```json
{
  "path": ".",
  "depth": 2,
  "type": "file",
  "newer_than": 300,
  "no_hidden": true,
  "context": true
}
```

Only `path` is required. The MCP tool preserves the same nested output, filtering, symlink, and context semantics as the CLI.

See [docs/mcp.md](docs/mcp.md) for client configuration examples.

## Development

```sh
make test
make trellis-lint
make build
make bench
```

Build a stamped local binary:

```sh
make build VERSION=0.1.0
./bin/loupe --version
```

Build a release artifact for the current platform:

```sh
make release VERSION=0.1.0
```

Cross-build by setting `GOOS` and `GOARCH`:

```sh
make release VERSION=0.1.0 GOOS=linux GOARCH=amd64
make release VERSION=0.1.0 GOOS=darwin GOARCH=arm64
```

Once the module has a tagged public release, install with:

```sh
go install github.com/norlinga/loupe@latest
```
