package mcp

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/norlinga/loupe/internal/schema"
)

func TestCallObserveReturnsLoupeJSONText(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "main.go")
	writeFile(t, path, "package main")

	result, err := CallObserve(json.RawMessage(`{"path":"` + filepath.ToSlash(path) + `"}`))
	if err != nil {
		t.Fatal(err)
	}
	text := resultText(t, result)
	var node schema.Node
	if err := json.Unmarshal([]byte(text), &node); err != nil {
		t.Fatal(err)
	}
	if node.Path != path || node.Type != schema.TypeFile {
		t.Fatalf("node = %#v", node)
	}
}

func TestCallObserveRequiresPath(t *testing.T) {
	_, err := CallObserve(json.RawMessage(`{}`))
	if err == nil || !strings.Contains(err.Error(), "path is required") {
		t.Fatalf("err = %v, want path required", err)
	}
}

func TestServeHandlesInitializeToolsListAndCall(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "main.go")
	writeFile(t, path, "package main")
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"loupe_observe","arguments":{"path":"` + filepath.ToSlash(path) + `"}}}`,
		"",
	}, "\n")
	var out bytes.Buffer

	if err := Serve(strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	responses := decodeResponses(t, out.String())
	if len(responses) != 3 {
		t.Fatalf("response count = %d, want 3: %s", len(responses), out.String())
	}
	if responses[0].Error != nil {
		t.Fatalf("initialize error = %#v", responses[0].Error)
	}
	if responses[1].Error != nil {
		t.Fatalf("tools/list error = %#v", responses[1].Error)
	}
	if !strings.Contains(string(responses[1].Result), "loupe_observe") {
		t.Fatalf("tools/list result = %s", responses[1].Result)
	}
	if !strings.Contains(string(responses[1].Result), "loupe_output_schema") {
		t.Fatalf("tools/list result = %s", responses[1].Result)
	}
	if responses[2].Error != nil {
		t.Fatalf("tools/call error = %#v", responses[2].Error)
	}
	if !strings.Contains(string(responses[2].Result), `main.go`) {
		t.Fatalf("tools/call result = %s", responses[2].Result)
	}
}

func TestCallSchemaTools(t *testing.T) {
	for _, tc := range []struct {
		name string
		want string
	}{
		{name: "loupe_output_schema", want: `"title": "loupe output"`},
		{name: "loupe_notes_schema", want: `"title": "loupe agent notes"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			params := json.RawMessage(`{"name":"` + tc.name + `","arguments":{}}`)
			result, err := handleToolCall(params)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(resultText(t, result), tc.want) {
				t.Fatalf("result = %#v, want %s", result, tc.want)
			}
		})
	}
}

func TestServeReturnsParseErrorForMalformedJSON(t *testing.T) {
	var out bytes.Buffer
	input := `{not json}` + "\n"

	if err := Serve(strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	responses := decodeResponses(t, out.String())
	if len(responses) != 1 {
		t.Fatalf("response count = %d, want 1", len(responses))
	}
	if responses[0].Error == nil || responses[0].Error.Code != -32700 {
		t.Fatalf("error = %#v, want parse error -32700", responses[0].Error)
	}
	if string(responses[0].ID) != "null" {
		t.Fatalf("id = %s, want null", responses[0].ID)
	}
}

func TestServeReturnsMethodNotFound(t *testing.T) {
	var out bytes.Buffer
	input := `{"jsonrpc":"2.0","id":1,"method":"unknown","params":{}}` + "\n"

	if err := Serve(strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	responses := decodeResponses(t, out.String())
	if len(responses) != 1 {
		t.Fatalf("response count = %d, want 1", len(responses))
	}
	if responses[0].Error == nil || responses[0].Error.Code != -32601 {
		t.Fatalf("error = %#v, want method not found", responses[0].Error)
	}
}

func resultText(t *testing.T, result map[string]any) string {
	t.Helper()
	content, ok := result["content"].([]map[string]any)
	if !ok || len(content) != 1 {
		t.Fatalf("content = %#v", result["content"])
	}
	text, ok := content[0]["text"].(string)
	if !ok {
		t.Fatalf("text = %#v", content[0]["text"])
	}
	return text
}

type testResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

func decodeResponses(t *testing.T, output string) []testResponse {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	responses := make([]testResponse, 0, len(lines))
	for _, line := range lines {
		var resp testResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			t.Fatal(err)
		}
		responses = append(responses, resp)
	}
	return responses
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
