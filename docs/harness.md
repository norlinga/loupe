# Harness Quickstart

This guide is for wiring `loupe` into an agent or MCP harness.

## 1. Build or Install

From the repository:

```sh
make build VERSION=dev
```

The binary is written to `bin/loupe`.

Install into a prefix:

```sh
make install PREFIX="$HOME/.local"
```

After a public tagged release exists:

```sh
go install github.com/norlinga/loupe@latest
```

## 2. Verify Binary-Discoverable Contracts

```sh
loupe --help
loupe --schema
loupe --notes-schema
```

Agents can use these commands even when repository docs are not installed beside the binary.

## 3. Configure MCP

Use an absolute path to the binary:

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

Available MCP tools:

- `loupe_observe`
- `loupe_output_schema`
- `loupe_notes_schema`

## 4. Install the Codex Skill

Copy the repository skill into Codex's skill directory:

```sh
mkdir -p "${CODEX_HOME:-$HOME/.codex}/skills"
cp -R skills/loupe "${CODEX_HOME:-$HOME/.codex}/skills/loupe"
```

The skill tells agents to prefer `loupe` over `ls`, `find`, and `stat` when observing local filesystem state.

## 5. Minimal Agent Instruction

For harnesses that do not support skills, add this instruction:

```text
When observing the local filesystem, prefer `loupe <path>` over `ls`, `find`, and `stat`.
Use `loupe --schema` to inspect the output contract and `loupe --notes-schema` for `.loupe/notes.json`.
Use `loupe --context` when project-level context or prior agent notes may matter.
Pipe loupe JSON to `jq` for filtering and transformation.
```
