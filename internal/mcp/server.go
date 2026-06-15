package mcp

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	loupecontext "github.com/norlinga/loupe/internal/context"
	"github.com/norlinga/loupe/internal/docs"
	"github.com/norlinga/loupe/internal/observe"
	"github.com/norlinga/loupe/internal/version"
)

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type observeArgs struct {
	Path      string `json:"path"`
	Depth     *int   `json:"depth,omitempty"`
	Type      string `json:"type,omitempty"`
	NewerThan int64  `json:"newer_than,omitempty"`
	NoHidden  bool   `json:"no_hidden,omitempty"`
	Context   bool   `json:"context,omitempty"`
}

func Serve(stdin io.Reader, stdout io.Writer) error {
	scanner := bufio.NewScanner(stdin)
	writer := bufio.NewWriter(stdout)
	defer writer.Flush()
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			if err := writeResponse(writer, response{
				JSONRPC: "2.0",
				ID:      json.RawMessage("null"),
				Error:   &rpcError{Code: -32700, Message: "parse error"},
			}); err != nil {
				return err
			}
			continue
		}
		if len(req.ID) == 0 {
			continue
		}
		if err := writeResponse(writer, handle(req)); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func handle(req request) response {
	resp := response{JSONRPC: "2.0", ID: req.ID}
	switch req.Method {
	case "initialize":
		resp.Result = initializeResult()
	case "tools/list":
		resp.Result = toolsListResult()
	case "tools/call":
		result, err := handleToolCall(req.Params)
		if err != nil {
			resp.Error = &rpcError{Code: -32602, Message: err.Error()}
			return resp
		}
		resp.Result = result
	default:
		resp.Error = &rpcError{Code: -32601, Message: "method not found"}
	}
	return resp
}

func initializeResult() map[string]any {
	return map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    "loupe",
			"version": version.String(),
		},
	}
}

func toolsListResult() map[string]any {
	return map[string]any{
		"tools": []map[string]any{
			{
				"name":        "loupe_observe",
				"description": "Observe a local filesystem path and return typed loupe JSON.",
				"inputSchema": map[string]any{
					"type":     "object",
					"required": []string{"path"},
					"properties": map[string]any{
						"path":       map[string]any{"type": "string", "description": "Local filesystem path to observe."},
						"depth":      map[string]any{"type": "integer", "description": "Recursion depth. Defaults to 1 for directories and 0 for files."},
						"type":       map[string]any{"type": "string", "enum": []string{"file", "dir", "directory", "symlink"}},
						"newer_than": map[string]any{"type": "integer", "description": "Only include entries modified in the last N seconds."},
						"no_hidden":  map[string]any{"type": "boolean", "description": "Exclude dotfile entries."},
						"context":    map[string]any{"type": "boolean", "description": "Include project context."},
					},
				},
			},
			{
				"name":        "loupe_output_schema",
				"description": "Return the JSON Schema for loupe observation output.",
				"inputSchema": emptyInputSchema(),
			},
			{
				"name":        "loupe_notes_schema",
				"description": "Return the JSON Schema for .loupe/notes.json.",
				"inputSchema": emptyInputSchema(),
			},
		},
	}
}

func emptyInputSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"properties":           map[string]any{},
		"additionalProperties": false,
	}
}

func handleToolCall(params json.RawMessage) (map[string]any, error) {
	var call toolCallParams
	if err := json.Unmarshal(params, &call); err != nil {
		return nil, err
	}
	switch call.Name {
	case "loupe_observe":
		return CallObserve(call.Arguments)
	case "loupe_output_schema":
		return toolText(docs.OutputSchema()), nil
	case "loupe_notes_schema":
		return toolText(docs.NotesSchema()), nil
	default:
		return nil, fmt.Errorf("unknown tool %q", call.Name)
	}
}

func CallObserve(raw json.RawMessage) (map[string]any, error) {
	var args observeArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	if args.Path == "" {
		return nil, errors.New("path is required")
	}
	entryType, err := observe.ParseType(args.Type)
	if err != nil {
		return nil, err
	}
	depth := -1
	if args.Depth != nil {
		depth = *args.Depth
	}
	node, err := observe.Observe(args.Path, observe.Options{
		Depth:     depth,
		Type:      entryType,
		NewerThan: time.Duration(args.NewerThan) * time.Second,
		NoHidden:  args.NoHidden,
	})
	if err != nil {
		return toolError(err.Error()), nil
	}
	if args.Context {
		loupecontext.Enrich(node, loupecontext.Options{NoHidden: args.NoHidden})
	}
	data, err := json.MarshalIndent(node, "", "  ")
	if err != nil {
		return nil, err
	}
	return toolText(string(data)), nil
}

func toolText(text string) map[string]any {
	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
	}
}

func toolError(message string) map[string]any {
	result := toolText(message)
	result["isError"] = true
	return result
}

func writeResponse(w *bufio.Writer, resp response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	if err := w.WriteByte('\n'); err != nil {
		return err
	}
	return w.Flush()
}
