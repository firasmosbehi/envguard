// Package watch provides file system watching for EnvGuard validation.
package watch

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors files and triggers validation.
type Watcher struct {
	paths      []string
	onChange   func() error
	debounce   time.Duration
	cmdSuccess string
	cmdFail    string
	quiet      bool
}

// Options configures the watcher.
type Options struct {
	Paths      []string
	Debounce   time.Duration
	CmdSuccess string // command to run on validation success
	CmdFail    string // command to run on validation failure
	Quiet      bool
}

// New creates a new Watcher.
func New(opts Options) *Watcher {
	debounce := opts.Debounce
	if debounce == 0 {
		debounce = 300 * time.Millisecond
	}
	return &Watcher{
		paths:      opts.Paths,
		debounce:   debounce,
		cmdSuccess: opts.CmdSuccess,
		cmdFail:    opts.CmdFail,
		quiet:      opts.Quiet,
	}
}

// SetCallback sets the function to call when files change.
func (w *Watcher) SetCallback(fn func() error) {
	w.onChange = fn
}

// Run starts watching and blocks until the context is cancelled.
func (w *Watcher) Run(ctx context.Context) error {
	if w.onChange == nil {
		return fmt.Errorf("no callback set")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	// Add files and their directories
	watchedDirs := make(map[string]bool)
	for _, path := range w.paths {
		abs, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		info, err := os.Stat(abs)
		if err != nil {
			continue
		}
		if info.IsDir() {
			watcher.Add(abs)
			watchedDirs[abs] = true
		} else {
			watcher.Add(abs)
			dir := filepath.Dir(abs)
			if !watchedDirs[dir] {
				watcher.Add(dir)
				watchedDirs[dir] = true
			}
		}
	}

	if len(watchedDirs) == 0 {
		return fmt.Errorf("no valid paths to watch")
	}

	// Debounce timer
	var mu sync.Mutex
	var timer *time.Timer
	trigger := make(chan struct{}, 1)

	// Run initial validation
	w.runValidation()

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if !w.isWatched(event.Name) {
				continue
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
				mu.Lock()
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(w.debounce, func() {
					select {
					case trigger <- struct{}{}:
					default:
					}
				})
				mu.Unlock()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			if !w.quiet {
				fmt.Fprintf(os.Stderr, "Watcher error: %v\n", err)
			}
		case <-trigger:
			w.runValidation()
		}
	}
}

func (w *Watcher) isWatched(path string) bool {
	for _, p := range w.paths {
		absP, _ := filepath.Abs(p)
		absPath, _ := filepath.Abs(path)
		if strings.HasPrefix(absPath, absP) || absPath == absP {
			return true
		}
	}
	return false
}

func (w *Watcher) runValidation() {
	if !w.quiet {
		clearScreen()
		fmt.Println("┌─────────────────────────────────────────┐")
		fmt.Println("│      EnvGuard Watch Mode                │")
		fmt.Println("├─────────────────────────────────────────┤")
		fmt.Printf("│  Watching: %-28s │\n", strings.Join(w.paths, ", "))
		fmt.Println("└─────────────────────────────────────────┘")
		fmt.Println()
	}

	err := w.onChange()
	if err != nil {
		if w.cmdFail != "" {
			w.runCommand(w.cmdFail)
		}
	} else {
		if w.cmdSuccess != "" {
			w.runCommand(w.cmdSuccess)
		}
	}
}

func (w *Watcher) runCommand(cmdStr string) {
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil && !w.quiet {
		fmt.Fprintf(os.Stderr, "Command failed: %v\n", err)
	}
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
