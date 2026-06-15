# Changelog

## Unreleased

### Added

- Agent-first filesystem observation CLI that emits typed JSON by default.
- Nested directory recursion with stable lexical entry ordering.
- Filters for depth, type, recent modification time, and hidden entries.
- Symlink support, including broken symlink target reporting.
- Project context with VCS, project type, recently modified files, and `.loupe/notes.json` agent notes.
- Embedded output schema via `loupe --schema`.
- Embedded agent notes schema via `loupe --notes-schema`.
- Stdio MCP server via `loupe --mcp`.
- MCP tools for observation and schema discovery:
  - `loupe_observe`
  - `loupe_output_schema`
  - `loupe_notes_schema`
- Trellis sidecars for source-level contracts.
- Golden tests, pathological filename tests, binary-level tests, MCP protocol tests, and observation benchmarks.
- Release workflow for tag-triggered Linux/macOS amd64/arm64 artifacts.
- Reusable Codex skill at `skills/loupe`.
