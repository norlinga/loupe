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

## Flags

- `--depth N`: recurse `N` levels deep. Defaults to `1` for directories and `0` for files.
- `--type file|dir|directory|symlink`: filter emitted entries by type.
- `--newer-than N`: emit entries modified in the last `N` seconds.
- `--no-hidden`: exclude dotfile entries.
- `--context`: include project context such as VCS, project type, recently modified files, and agent notes.
- `--human`: print a minimal human-readable tree.
- `--version`: print the build version and exit.

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
