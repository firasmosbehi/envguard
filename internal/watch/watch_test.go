package watch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatcher_Callback(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, ".env")
	os.WriteFile(testFile, []byte("KEY1=value1\n"), 0644)

	var callCount atomic.Int32
	w := New(Options{
		Paths:    []string{testFile},
		Debounce: 100 * time.Millisecond,
	})
	w.SetCallback(func() error {
		callCount.Add(1)
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start watcher in background
	done := make(chan error, 1)
	go func() {
		done <- w.Run(ctx)
	}()

	// Wait for initial run
	time.Sleep(200 * time.Millisecond)
	if callCount.Load() != 1 {
		t.Fatalf("expected 1 initial call, got %d", callCount.Load())
	}

	// Trigger a change
	os.WriteFile(testFile, []byte("KEY1=value2\n"), 0644)
	time.Sleep(300 * time.Millisecond)

	if callCount.Load() < 2 {
		t.Fatalf("expected at least 2 calls, got %d", callCount.Load())
	}

	cancel()
	<-done
}

func TestWatcher_Debounce(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, ".env")
	os.WriteFile(testFile, []byte("KEY1=value1\n"), 0644)

	var callCount atomic.Int32
	w := New(Options{
		Paths:    []string{testFile},
		Debounce: 200 * time.Millisecond,
	})
	w.SetCallback(func() error {
		callCount.Add(1)
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- w.Run(ctx)
	}()

	// Wait for initial
	time.Sleep(200 * time.Millisecond)
	initial := callCount.Load()

	// Rapid changes should be debounced into a single call
	for i := 0; i < 5; i++ {
		os.WriteFile(testFile, []byte(fmt.Sprintf("KEY1=value%d\n", i)), 0644)
		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(400 * time.Millisecond)
	if callCount.Load() > initial+2 {
		t.Fatalf("expected debounced calls (%d initial + at most 2), got %d", initial, callCount.Load())
	}

	cancel()
	<-done
}

func TestWatcher_NoCallback(t *testing.T) {
	w := New(Options{Paths: []string{"."}})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := w.Run(ctx)
	if err == nil || err.Error() != "no callback set" {
		t.Fatalf("expected 'no callback set' error, got %v", err)
	}
}

func TestWatcher_Quiet(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, ".env")
	os.WriteFile(testFile, []byte("KEY=value\n"), 0644)

	w := New(Options{
		Paths:    []string{testFile},
		Debounce: 100 * time.Millisecond,
		Quiet:    true,
	})
	w.SetCallback(func() error { return nil })

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- w.Run(ctx)
	}()

	time.Sleep(300 * time.Millisecond)
	cancel()
	<-done
	// Should complete without printing
}

func TestIsWatched(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, ".env")
	os.WriteFile(testFile, []byte("KEY=value\n"), 0644)

	w := New(Options{Paths: []string{testFile}})
	if !w.isWatched(testFile) {
		t.Error("expected testFile to be watched")
	}
	if w.isWatched("/nonexistent/path") {
		t.Error("expected nonexistent path to not be watched")
	}
}

func TestWatcher_NoValidPaths(t *testing.T) {
	w := New(Options{Paths: []string{"/nonexistent/path/.env"}})
	w.SetCallback(func() error { return nil })

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := w.Run(ctx)
	if err == nil || err.Error() != "no valid paths to watch" {
		t.Fatalf("expected 'no valid paths to watch' error, got %v", err)
	}
}

func TestWatcher_InvalidPaths(t *testing.T) {
	// Mix of valid and invalid paths
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, ".env")
	os.WriteFile(testFile, []byte("KEY=value\n"), 0644)

	w := New(Options{
		Paths: []string{"/nonexistent/path/.env", testFile},
	})
	w.SetCallback(func() error { return nil })

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- w.Run(ctx)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()
	<-done
}

func TestWatcher_WatcherErrorQuiet(t *testing.T) {
	// fsnotify errors are hard to trigger, but we can at least test quiet mode
	// by creating a valid watcher and letting it run briefly
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, ".env")
	os.WriteFile(testFile, []byte("KEY=value\n"), 0644)

	w := New(Options{
		Paths:    []string{testFile},
		Debounce: 50 * time.Millisecond,
		Quiet:    true,
	})
	w.SetCallback(func() error { return nil })

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- w.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	<-done
}

func TestRunValidation_Quiet(t *testing.T) {
	callCount := 0
	w := New(Options{
		Paths:      []string{"."},
		Quiet:      true,
		CmdSuccess: "echo success",
	})
	w.SetCallback(func() error {
		callCount++
		return nil
	})

	w.runValidation()
	if callCount != 1 {
		t.Fatalf("expected 1 call, got %d", callCount)
	}
}

func TestRunValidation_CallbackError(t *testing.T) {
	callCount := 0
	w := New(Options{
		Paths:   []string{"."},
		Quiet:   true,
		CmdFail: "echo fail",
	})
	w.SetCallback(func() error {
		callCount++
		return fmt.Errorf("validation failed")
	})

	w.runValidation()
	if callCount != 1 {
		t.Fatalf("expected 1 call, got %d", callCount)
	}
}

func TestRunCommand_Success(t *testing.T) {
	w := New(Options{Paths: []string{"."}, Quiet: true})
	w.runCommand("echo hello")
}

func TestRunCommand_Failure(t *testing.T) {
	w := New(Options{Paths: []string{"."}, Quiet: true})
	// Run a command that will fail
	w.runCommand("false")
}

func TestRunCommand_FailureNotQuiet(t *testing.T) {
	w := New(Options{Paths: []string{"."}, Quiet: false})
	// Run a command that will fail; should print error
	w.runCommand("false")
}

func TestRunCommand_Empty(t *testing.T) {
	w := New(Options{Paths: []string{"."}})
	w.runCommand("")
	// Should not panic
}

func TestIsWatched_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	os.MkdirAll(subDir, 0755)

	w := New(Options{Paths: []string{tmpDir}})
	if !w.isWatched(subDir) {
		t.Error("expected subdirectory to be watched when parent dir is in paths")
	}
	if !w.isWatched(filepath.Join(subDir, ".env")) {
		t.Error("expected file in subdirectory to be watched")
	}
}
