// Package lsp provides a Language Server Protocol implementation for EnvGuard.
package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/envguard/envguard/internal/dotenv"
	"github.com/envguard/envguard/internal/schema"
	"github.com/envguard/envguard/internal/validator"
)

// Message represents an LSP JSON-RPC message.
type Message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents an LSP JSON-RPC response.
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      *int        `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *LSPError   `json:"error,omitempty"`
}

type LSPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// PublishDiagnosticsParams represents LSP diagnostic notification.
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Version     int          `json:"version,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"`
	Code     string `json:"code,omitempty"`
	Source   string `json:"source"`
	Message  string `json:"message"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Server handles LSP communication.
type Server struct {
	reader *bufio.Reader
	writer io.Writer
	schema *schema.Schema
}

// NewServer creates a new LSP server.
func NewServer(r io.Reader, w io.Writer) *Server {
	return &Server{
		reader: bufio.NewReader(r),
		writer: w,
	}
}

// Run starts the LSP server and processes messages.
func (s *Server) Run() error {
	for {
		msg, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch msg.Method {
		case "initialize":
			s.writeResponse(msg.ID, map[string]interface{}{
				"capabilities": map[string]interface{}{
					"textDocumentSync": map[string]interface{}{
						"openClose": true,
						"change":    1, // full
					},
				},
				"serverInfo": map[string]interface{}{
					"name":    "envguard-lsp",
					"version": "2.0.0",
				},
			})
		case "initialized":
			// no-op
		case "textDocument/didOpen", "textDocument/didChange":
			var params struct {
				TextDocument struct {
					URI     string `json:"uri"`
					Version int    `json:"version"`
					Text    string `json:"text"`
				} `json:"textDocument"`
			}
			if err := json.Unmarshal(msg.Params, &params); err == nil {
				s.validateDocument(params.TextDocument.URI, params.TextDocument.Version, params.TextDocument.Text)
			}
		case "shutdown":
			s.writeResponse(msg.ID, nil)
		case "exit":
			return nil
		}
	}
}

func (s *Server) readMessage() (*Message, error) {
	var contentLength int
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "Content-Length: ") {
			fmt.Sscanf(line, "Content-Length: %d", &contentLength)
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("no content length")
	}

	body := make([]byte, contentLength)
	if _, err := io.ReadFull(s.reader, body); err != nil {
		return nil, err
	}

	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, err
	}
	msg.JSONRPC = "2.0"
	return &msg, nil
}

func (s *Server) writeResponse(id *int, result interface{}) {
	resp := Response{JSONRPC: "2.0", ID: id, Result: result}
	data, _ := json.Marshal(resp)
	fmt.Fprintf(s.writer, "Content-Length: %d\r\n\r\n%s", len(data), data)
}

func (s *Server) sendNotification(method string, params interface{}) {
	msg := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	data, _ := json.Marshal(msg)
	fmt.Fprintf(s.writer, "Content-Length: %d\r\n\r\n%s", len(data), data)
}

func (s *Server) validateDocument(uri string, version int, text string) {
	// Load schema from workspace
	if s.schema == nil {
		schemaPath := findSchemaForURI(uri)
		if schemaPath != "" {
			sch, err := schema.Parse(schemaPath)
			if err == nil {
				s.schema = sch
			}
		}
	}

	if s.schema == nil {
		return
	}

	// Parse the document text as .env
	vars := parseEnvText(text)
	result := validator.Validate(s.schema, vars, false, "")

	var diagnostics []Diagnostic
	for _, err := range result.Errors {
		line := findLineForKey(text, err.Key)
		diagnostics = append(diagnostics, Diagnostic{
			Range: Range{
				Start: Position{Line: line, Character: 0},
				End:   Position{Line: line, Character: len(err.Key)},
			},
			Severity: severityToLSP(string(err.Severity)),
			Code:     err.Rule,
			Source:   "envguard",
			Message:  err.Message,
		})
	}

	s.sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Version:     version,
		Diagnostics: diagnostics,
	})
}

func parseEnvText(text string) map[string]string {
	// Write to temp file and use dotenv parser
	tmp, _ := os.CreateTemp("", "envguard-lsp-*.env")
	defer os.Remove(tmp.Name())
	tmp.WriteString(text)
	tmp.Close()
	parsed, _ := dotenv.Parse(tmp.Name())
	return parsed
}

func findLineForKey(text, key string) int {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, key+"=") {
			return i
		}
	}
	return 0
}

func findSchemaForURI(uri string) string {
	path := strings.TrimPrefix(uri, "file://")
	dir := filepath.Dir(path)
	for {
		schemaPath := filepath.Join(dir, "envguard.yaml")
		if _, err := os.Stat(schemaPath); err == nil {
			return schemaPath
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func severityToLSP(sev string) int {
	switch sev {
	case "error":
		return 1 // Error
	case "warn":
		return 2 // Warning
	case "info":
		return 3 // Information
	default:
		return 1
	}
}
