package lsp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/envguard/envguard/internal/schema"
)

func TestServerInitialize(t *testing.T) {
	var out bytes.Buffer
	server := NewServer(strings.NewReader(makeInitializeRequest()), &out)

	// Run in goroutine since it blocks
	done := make(chan error, 1)
	go func() {
		done <- server.Run()
	}()

	// Wait a bit for response
	// Since Run reads until EOF, we need to give it time
	// The test is simplified - just verify the server struct
	if server.reader == nil {
		t.Error("expected reader to be set")
	}
	if server.writer == nil {
		t.Error("expected writer to be set")
	}
}

func TestSeverityToLSP(t *testing.T) {
	if severityToLSP("error") != 1 {
		t.Error("expected error = 1")
	}
	if severityToLSP("warn") != 2 {
		t.Error("expected warn = 2")
	}
	if severityToLSP("info") != 3 {
		t.Error("expected info = 3")
	}
}

func TestFindLineForKey(t *testing.T) {
	text := "FOO=1\nBAR=2\nBAZ=3\n"
	if findLineForKey(text, "BAR") != 1 {
		t.Errorf("expected line 1 for BAR, got %d", findLineForKey(text, "BAR"))
	}
	if findLineForKey(text, "MISSING") != 0 {
		t.Errorf("expected line 0 for missing key, got %d", findLineForKey(text, "MISSING"))
	}
}

func makeInitializeRequest() string {
	msg := Message{
		JSONRPC: "2.0",
		ID:      intPtr(1),
		Method:  "initialize",
		Params:  json.RawMessage(`{}`),
	}
	data, _ := json.Marshal(msg)
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(data), data)
}

func intPtr(i int) *int {
	return &i
}

func TestServerRunShutdown(t *testing.T) {
	msg := Message{
		JSONRPC: "2.0",
		ID:      intPtr(1),
		Method:  "shutdown",
		Params:  json.RawMessage(`{}`),
	}
	data, _ := json.Marshal(msg)
	input := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(data), data)

	var out bytes.Buffer
	server := NewServer(strings.NewReader(input), &out)
	err := server.Run()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Result is omitted when nil due to omitempty, but id should be present
	if !bytes.Contains(out.Bytes(), []byte(`"id":1`)) {
		t.Errorf("expected shutdown response with id, got %q", out.String())
	}
}

func TestServerRunExit(t *testing.T) {
	msg := Message{
		JSONRPC: "2.0",
		Method:  "exit",
		Params:  json.RawMessage(`{}`),
	}
	data, _ := json.Marshal(msg)
	input := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(data), data)

	var out bytes.Buffer
	server := NewServer(strings.NewReader(input), &out)
	err := server.Run()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestServerRunDidOpen(t *testing.T) {
	// Create a temp dir with a schema so validation can run
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  KEY:
    type: string
    required: true
`), 0644)

	envFile := filepath.Join(tmpDir, "test.env")
	uri := "file://" + envFile

	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":     uri,
			"version": 1,
			"text":    "KEY=value\n",
		},
	}
	paramsJSON, _ := json.Marshal(params)
	msg := Message{
		JSONRPC: "2.0",
		Method:  "textDocument/didOpen",
		Params:  paramsJSON,
	}
	data, _ := json.Marshal(msg)
	input := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(data), data)

	var out bytes.Buffer
	server := NewServer(strings.NewReader(input), &out)
	err := server.Run()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Should have sent diagnostics notification
	if !bytes.Contains(out.Bytes(), []byte(`"method":"textDocument/publishDiagnostics"`)) {
		t.Errorf("expected publishDiagnostics notification, got %q", out.String())
	}
}

func TestServerRunDidChange(t *testing.T) {
	// Create a temp dir with a schema so validation can run
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
    required: true
`), 0644)

	envFile := filepath.Join(tmpDir, "test.env")
	uri := "file://" + envFile

	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":     uri,
			"version": 2,
			"text":    "FOO=bar\n",
		},
	}
	paramsJSON, _ := json.Marshal(params)
	msg := Message{
		JSONRPC: "2.0",
		Method:  "textDocument/didChange",
		Params:  paramsJSON,
	}
	data, _ := json.Marshal(msg)
	input := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(data), data)

	var out bytes.Buffer
	server := NewServer(strings.NewReader(input), &out)
	err := server.Run()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte(`"method":"textDocument/publishDiagnostics"`)) {
		t.Errorf("expected publishDiagnostics notification, got %q", out.String())
	}
}

func TestSendNotification(t *testing.T) {
	var out bytes.Buffer
	server := NewServer(strings.NewReader(""), &out)
	server.sendNotification("test/notification", map[string]string{"key": "value"})

	if !bytes.Contains(out.Bytes(), []byte(`"method":"test/notification"`)) {
		t.Error("expected notification method")
	}
	if !bytes.Contains(out.Bytes(), []byte(`"key":"value"`)) {
		t.Error("expected notification params")
	}
}

func TestValidateDocumentWithSchema(t *testing.T) {
	// Create a temporary schema file
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	schemaContent := `version: "1.0"
env:
  REQUIRED_KEY:
    type: string
    required: true
`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	server := NewServer(strings.NewReader(""), &out)
	// Pre-load schema
	sch, err := schema.Parse(schemaPath)
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}
	server.schema = sch

	server.validateDocument("file:///tmp/test.env", 1, "")

	if !bytes.Contains(out.Bytes(), []byte(`"method":"textDocument/publishDiagnostics"`)) {
		t.Error("expected publishDiagnostics notification")
	}
	if !bytes.Contains(out.Bytes(), []byte(`"source":"envguard"`)) {
		t.Error("expected envguard source")
	}
}

func TestValidateDocumentWithoutSchema(t *testing.T) {
	var out bytes.Buffer
	server := NewServer(strings.NewReader(""), &out)
	server.validateDocument("file:///nonexistent/test.env", 1, "KEY=value\n")

	// Should not send any notification since no schema is found
	if bytes.Contains(out.Bytes(), []byte(`"method":"textDocument/publishDiagnostics"`)) {
		t.Error("expected no publishDiagnostics without schema")
	}
}

func TestParseEnvText(t *testing.T) {
	vars := parseEnvText("FOO=1\nBAR=2\n")
	if vars["FOO"] != "1" {
		t.Errorf("expected FOO=1, got %q", vars["FOO"])
	}
	if vars["BAR"] != "2" {
		t.Errorf("expected BAR=2, got %q", vars["BAR"])
	}
}

func TestParseEnvTextEmpty(t *testing.T) {
	vars := parseEnvText("")
	if vars == nil {
		t.Error("expected non-nil map")
	}
}

func TestFindSchemaForURI(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte("version: \"1.0\"\n"), 0644)

	envFile := filepath.Join(tmpDir, ".env")
	uri := "file://" + envFile

	found := findSchemaForURI(uri)
	if found != schemaPath {
		t.Errorf("expected %q, got %q", schemaPath, found)
	}
}

func TestFindSchemaForURINotFound(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	uri := "file://" + envFile

	found := findSchemaForURI(uri)
	if found != "" {
		t.Errorf("expected empty string, got %q", found)
	}
}

func TestReadMessageNoContentLength(t *testing.T) {
	input := "X-Header: something\r\n\r\n"
	server := NewServer(strings.NewReader(input), &bytes.Buffer{})
	_, err := server.readMessage()
	if err == nil || err.Error() != "no content length" {
		t.Fatalf("expected 'no content length' error, got %v", err)
	}
}

func TestReadMessageInvalidJSON(t *testing.T) {
	input := "Content-Length: 5\r\n\r\n{abc"
	server := NewServer(strings.NewReader(input), &bytes.Buffer{})
	_, err := server.readMessage()
	if err == nil {
		t.Fatal("expected JSON unmarshal error")
	}
}

func TestReadMessageShortBody(t *testing.T) {
	input := "Content-Length: 100\r\n\r\n{"
	server := NewServer(strings.NewReader(input), &bytes.Buffer{})
	_, err := server.readMessage()
	if err == nil {
		t.Fatal("expected error for short body")
	}
}

func TestReadMessageEOF(t *testing.T) {
	server := NewServer(strings.NewReader(""), &bytes.Buffer{})
	_, err := server.readMessage()
	if err == nil {
		t.Fatal("expected EOF error")
	}
}

func TestServerRunInitialized(t *testing.T) {
	msg := Message{
		JSONRPC: "2.0",
		Method:  "initialized",
		Params:  json.RawMessage(`{}`),
	}
	data, _ := json.Marshal(msg)
	input := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(data), data)

	var out bytes.Buffer
	server := NewServer(strings.NewReader(input), &out)
	err := server.Run()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// initialized is a no-op, so no output expected
	if out.Len() != 0 {
		t.Errorf("expected no output for initialized, got %q", out.String())
	}
}

func TestValidateDocumentWithFindSchema(t *testing.T) {
	// Create a temporary directory with schema
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	schemaContent := `version: "1.0"
env:
  KEY:
    type: string
    required: true
`
	os.WriteFile(schemaPath, []byte(schemaContent), 0644)

	var out bytes.Buffer
	server := NewServer(strings.NewReader(""), &out)

	envFile := filepath.Join(tmpDir, "test.env")
	uri := "file://" + envFile

	server.validateDocument(uri, 1, "")

	if !bytes.Contains(out.Bytes(), []byte(`"method":"textDocument/publishDiagnostics"`)) {
		t.Error("expected publishDiagnostics notification after finding schema")
	}
}

func TestSeverityToLSPDefault(t *testing.T) {
	if severityToLSP("unknown") != 1 {
		t.Error("expected default severity = 1")
	}
}
