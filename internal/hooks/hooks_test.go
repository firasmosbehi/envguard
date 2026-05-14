package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallAndUninstall(t *testing.T) {
	// Create a fake git repo
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git", "hooks")
	os.MkdirAll(gitDir, 0755)

	// Override findGitDir by changing into the temp dir
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	// Install pre-commit hook
	if err := Install(Options{HookType: "pre-commit", Force: false, Command: "envguard validate"}); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	hookPath := filepath.Join(gitDir, "pre-commit")
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("hook not found: %v", err)
	}
	if !strings.Contains(string(data), "envguard validate") {
		t.Errorf("hook missing envguard command: %s", string(data))
	}

	// Install without force should fail
	if err := Install(Options{HookType: "pre-commit", Force: false}); err == nil {
		t.Error("expected error when hook exists without --force")
	}

	// Status
	isEnvGuard, content, err := Status("pre-commit")
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if !isEnvGuard {
		t.Error("expected hook to be recognized as envguard hook")
	}
	if !strings.Contains(content, "envguard") {
		t.Error("expected content to contain envguard")
	}

	// Uninstall
	if err := Uninstall("pre-commit"); err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}
	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Error("hook should be removed")
	}
}

func TestUnsupportedHookType(t *testing.T) {
	if err := Install(Options{HookType: "post-commit"}); err == nil {
		t.Error("expected error for unsupported hook type")
	}
}

func TestNotInGitRepo(t *testing.T) {
	// Change to a temp dir without .git
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	if err := Install(Options{HookType: "pre-commit"}); err == nil {
		t.Error("expected error when not in git repo")
	}
}
