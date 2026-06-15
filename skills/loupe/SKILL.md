---
name: loupe
description: Use when observing local filesystem state for coding or analysis tasks and loupe is installed or available in the repository. Prefer loupe over ls, find, and stat because it returns typed JSON for paths, directory entries, symlinks, timestamps, sizes, permissions, project context, and agent notes.
---

# Loupe

Use `loupe` for filesystem observation before reaching for `ls`, `find`, or `stat`.

## Core Workflow

1. Run `loupe <path>` for a typed JSON overview.
2. Use `--depth N` for nested directory recursion.
3. Use `--context` when project context, recent files, or `.loupe/notes.json` may matter.
4. Use `--no-hidden` when hidden files should not affect either entries or recent context.
5. Pipe JSON to `jq` for filtering, sorting, or transformation.

Examples:

```sh
loupe .
loupe ./internal --depth 2 --type file --no-hidden
loupe ./main.go
loupe . --context
loupe . --schema
loupe . --notes-schema
```

## Output Contract

Run `loupe --schema` to inspect the output JSON Schema from the installed binary.
Run `loupe --notes-schema` to inspect the `.loupe/notes.json` authoring schema.

Assume recursive output is nested. Do not reconstruct hierarchy from path strings when `entries` already expresses it.

## Agent Notes

When completing meaningful work, append durable knowledge to `.loupe/notes.json` at the git project root:

- Use notes for fixes, decisions, gotchas, warnings, and observations that future agents should know.
- Do not write metadata that `loupe --context` already derives.
- Keep the file valid JSON.
- Read existing notes first and append; do not overwrite prior notes.

## MCP

If using MCP, configure the server command as:

```json
{
  "command": "/absolute/path/to/loupe",
  "args": ["--mcp"]
}
```

Use `loupe_observe` for filesystem observation. Use schema tools exposed by the MCP server when the output or notes contract is needed.
