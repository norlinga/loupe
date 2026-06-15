# MCP Integration

`loupe --mcp` starts a stdio MCP server.

## Generic Client Configuration

Use the built binary:

```json
{
  "mcpServers": {
    "loupe": {
      "command": "/absolute/path/to/loupe",
      "args": ["--mcp"]
    }
  }
}
```

During local development, point at the workspace build:

```json
{
  "mcpServers": {
    "loupe": {
      "command": "/home/aaron/code/github.com/norlinga/loupe/bin/loupe",
      "args": ["--mcp"]
    }
  }
}
```

## Tool

Tools:

- `loupe_observe`: observes a local filesystem path and returns loupe JSON as MCP text content.
- `loupe_output_schema`: returns the JSON Schema for loupe output.
- `loupe_notes_schema`: returns the JSON Schema for `.loupe/notes.json`.

## `loupe_observe`

Input schema:

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

Only `path` is required.

Fields:

- `path`: local filesystem path to observe.
- `depth`: recursion depth. Defaults to `1` for directories and `0` for files.
- `type`: optional emitted-entry filter: `file`, `dir`, `directory`, or `symlink`.
- `newer_than`: only include entries modified in the last `N` seconds.
- `no_hidden`: exclude hidden entries and hidden recent context paths.
- `context`: include project context.

The observe tool preserves the same output contract documented in [schema.md](schema.md).

## Example JSON-RPC Call

```json
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"loupe_observe","arguments":{"path":".","depth":1,"context":true}}}
```
