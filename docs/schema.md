# loupe Output Schema

`loupe` emits one JSON object representing the observed path. Directory recursion is nested: child directories may contain their own `entries` arrays up to the requested depth.

The machine-readable schema is [loupe.schema.json](loupe.schema.json).

Installed binaries also print the schema directly:

```sh
loupe --schema
```

## Node

Required fields:

- `path`: absolute filesystem path.
- `name`: base filename.
- `type`: one of `file`, `directory`, `symlink`, or `other`.
- `size_bytes`: integer byte size.
- `modified_unix`: Unix epoch timestamp in seconds.
- `permissions`: plain octal permission string.

Optional fields:

- `entry_count`: number of emitted child entries for a directory.
- `entries`: nested child nodes for directory observations.
- `extension`: file extension for regular files when present.
- `target`: symlink target. Broken symlinks report the unresolved absolute target path.
- `context`: project context when requested with `--context`.

## Context

The `context` object may include:

- `vcs`: currently `git` when a nearest git root is detected.
- `project_type`: detected from marker files such as `go.mod`, `package.json`, `Cargo.toml`, or `pyproject.toml`.
- `recently_modified_secs`: recent-file window in seconds.
- `recently_modified`: project-root-relative file paths sorted newest first, using lexical order as a tie-breaker, capped to the configured limit.
- `agent_notes`: notes read from `.loupe/notes.json` at the nearest git project root.

The `.loupe/notes.json` authoring schema is [notes.schema.json](notes.schema.json). Installed binaries print it directly:

```sh
loupe --notes-schema
```

When `--no-hidden` is used, hidden entries are excluded from both the observed tree and `context.recently_modified`.

## Filtering

`--type` filters emitted entries, not traversal. If `--type file --depth 2` finds a matching file inside a directory, the containing directory remains in the nested output so the file is reachable.

`--newer-than` filters entries by modification time. Directory containers may still be emitted when they contain matching descendants.

## Errors

If the root path does not exist, or traversal encounters an unreadable descendant, `loupe` returns an error instead of partial JSON.
