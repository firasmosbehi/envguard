package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// safeBuffer is a thread-safe bytes.Buffer for use in concurrent tests.
type safeBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *safeBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *safeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func (s *safeBuffer) Bytes() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Bytes()
}

func (s *safeBuffer) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Len()
}

var _ io.Writer = (*safeBuffer)(nil)

// === runWatch coverage tests ===

func TestRunWatchQuietMode(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
`), 0644)
	os.WriteFile(envPath, []byte("FOO=bar\n"), 0644)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	opts := &watchOptions{
		schemaPath: schemaPath,
		envPaths:   []string{envPath},
		format:     "text",
		debounce:   50 * time.Millisecond,
		quiet:      true,
	}
	stdout := &safeBuffer{}
	stderr := &safeBuffer{}

	done := make(chan error, 1)
	go func() {
		done <- runWatch(stdout, stderr, opts)
	}()

	time.Sleep(300 * time.Millisecond)
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for runWatch")
	}

	if strings.Contains(stdout.String(), "Starting EnvGuard watch mode") {
		t.Errorf("quiet mode should suppress startup message")
	}
}

func TestRunWatchJSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
`), 0644)
	os.WriteFile(envPath, []byte("FOO=bar\n"), 0644)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	opts := &watchOptions{
		schemaPath: schemaPath,
		envPaths:   []string{envPath},
		format:     "json",
		debounce:   50 * time.Millisecond,
		quiet:      false,
	}
	stdout := &safeBuffer{}
	stderr := &safeBuffer{}

	done := make(chan error, 1)
	go func() {
		done <- runWatch(stdout, stderr, opts)
	}()

	time.Sleep(300 * time.Millisecond)
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for runWatch")
	}

	if !strings.Contains(stdout.String(), `"valid":true`) && !strings.Contains(stdout.String(), `"valid": true`) {
		t.Errorf("expected JSON output with valid=true, got: %s", stdout.String())
	}
}

func TestRunWatchCmdOnFail(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
    required: true
`), 0644)
	os.WriteFile(envPath, []byte(""), 0644)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	opts := &watchOptions{
		schemaPath: schemaPath,
		envPaths:   []string{envPath},
		format:     "text",
		debounce:   50 * time.Millisecond,
		quiet:      true,
		cmdFail:    "true",
	}
	stdout := &safeBuffer{}
	stderr := &safeBuffer{}

	done := make(chan error, 1)
	go func() {
		done <- runWatch(stdout, stderr, opts)
	}()

	time.Sleep(300 * time.Millisecond)
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for runWatch")
	}
}

func TestRunWatchInvalidSchemaQuiet(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte("invalid yaml: {"), 0644)
	os.WriteFile(envPath, []byte("FOO=bar\n"), 0644)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	opts := &watchOptions{
		schemaPath: schemaPath,
		envPaths:   []string{envPath},
		format:     "text",
		debounce:   50 * time.Millisecond,
		quiet:      true,
	}
	stdout := &safeBuffer{}
	stderr := &safeBuffer{}

	done := make(chan error, 1)
	go func() {
		done <- runWatch(stdout, stderr, opts)
	}()

	time.Sleep(300 * time.Millisecond)
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for runWatch")
	}

	if stderr.Len() > 0 {
		t.Errorf("quiet mode should suppress stderr, got: %s", stderr.String())
	}
}

func TestRunWatchInvalidEnvQuiet(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
`), 0644)
	os.WriteFile(envPath, []byte("=invalid\n"), 0644)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	opts := &watchOptions{
		schemaPath: schemaPath,
		envPaths:   []string{envPath},
		format:     "text",
		debounce:   50 * time.Millisecond,
		quiet:      true,
	}
	stdout := &safeBuffer{}
	stderr := &safeBuffer{}

	done := make(chan error, 1)
	go func() {
		done <- runWatch(stdout, stderr, opts)
	}()

	time.Sleep(300 * time.Millisecond)
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for runWatch")
	}

	if stderr.Len() > 0 {
		t.Errorf("quiet mode should suppress stderr, got: %s", stderr.String())
	}
}

// === LSP command execution test ===

func TestLSPCmdExecution(t *testing.T) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	stdinReader, stdinWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdin = stdinReader
	os.Stdout = stdoutWriter

	go func() {
		initMsg := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
		fmt.Fprintf(stdinWriter, "Content-Length: %d\r\n\r\n%s", len(initMsg), initMsg)

		exitMsg := `{"jsonrpc":"2.0","method":"exit"}`
		fmt.Fprintf(stdinWriter, "Content-Length: %d\r\n\r\n%s", len(exitMsg), exitMsg)
		stdinWriter.Close()
	}()

	// Drain stdout asynchronously to prevent blocking on pipe writes
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdoutReader.Read(buf)
			if n > 0 {
				// discard
			}
			if err != nil {
				return
			}
		}
	}()

	cmd := newLSPCmd()
	var stderrBuf bytes.Buffer
	cmd.SetErr(&stderrBuf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stderrBuf.String(), "EnvGuard LSP server starting") {
		t.Errorf("expected startup message, got: %s", stderrBuf.String())
	}
}

// === init.go error path tests ===

func TestInitCommandAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.WriteFile("envguard.yaml", []byte("version: \"1.0\"\n"), 0644)

	cmd := newInitCmd()
	cmd.SetArgs([]string{})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when envguard.yaml already exists")
	}
	if !strings.Contains(buf.String(), "already exists") {
		t.Errorf("expected already exists message, got: %s", buf.String())
	}
}

func TestInitCommandDefaultWriteError(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.Chmod(tmpDir, 0555)
	defer os.Chmod(tmpDir, 0755)

	cmd := newInitCmd()
	cmd.SetArgs([]string{})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for write failure")
	}
}

func TestGenerateConfigFileError(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.Chmod(tmpDir, 0555)
	defer os.Chmod(tmpDir, 0755)

	var stdout, stderr bytes.Buffer
	err := generateConfigFile(&stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for write failure")
	}
}

// === docs.go error path test ===

func TestDocsCommandWriteError(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
`), 0644)

	roDir := filepath.Join(tmpDir, "readonly")
	os.MkdirAll(roDir, 0555)
	defer os.Chmod(roDir, 0755)

	outputPath := filepath.Join(roDir, "out.md")

	cmd := newDocsCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--output", outputPath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for write failure")
	}
}
