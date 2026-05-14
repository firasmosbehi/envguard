package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUninstallUnsupportedHookType(t *testing.T) {
	if err := Uninstall("post-commit"); err == nil {
		t.Error("expected error for unsupported hook type")
	}
}

func TestUninstallNotInGitRepo(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	if err := Uninstall("pre-commit"); err == nil {
		t.Error("expected error when not in git repo")
	}
}

func TestUninstallNonExistentHook(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git", "hooks")
	os.MkdirAll(gitDir, 0755)

	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	err := Uninstall("pre-commit")
	if err == nil {
		t.Error("expected error when hook does not exist")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("expected 'does not exist' error, got: %v", err)
	}
}

func TestStatusUnsupportedHookType(t *testing.T) {
	_, _, err := Status("post-commit")
	if err == nil {
		t.Error("expected error for unsupported hook type")
	}
}

func TestStatusNotInGitRepo(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	_, _, err := Status("pre-commit")
	if err == nil {
		t.Error("expected error when not in git repo")
	}
}

func TestStatusHookDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git", "hooks")
	os.MkdirAll(gitDir, 0755)

	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	isEnvGuard, content, err := Status("pre-commit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isEnvGuard {
		t.Error("expected isEnvGuard to be false")
	}
	if content != "" {
		t.Errorf("expected empty content, got %q", content)
	}
}

func TestStatusHookExistsButNotEnvGuard(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git", "hooks")
	os.MkdirAll(gitDir, 0755)

	// Write a non-envguard hook
	hookPath := filepath.Join(gitDir, "pre-commit")
	os.WriteFile(hookPath, []byte("#!/bin/sh\necho 'custom hook'\n"), 0755)

	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	isEnvGuard, content, err := Status("pre-commit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isEnvGuard {
		t.Error("expected isEnvGuard to be false for non-envguard hook")
	}
	if content == "" {
		t.Error("expected non-empty content")
	}
}

func TestInstallPrePush(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git", "hooks")
	os.MkdirAll(gitDir, 0755)

	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	if err := Install(Options{HookType: "pre-push", Force: false}); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	hookPath := filepath.Join(gitDir, "pre-push")
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("hook not found: %v", err)
	}
	if !strings.Contains(string(data), "envguard validate --strict") {
		t.Errorf("hook missing default command: %s", string(data))
	}
}
