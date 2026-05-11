package run

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_absent(t *testing.T) {
	dir := t.TempDir()
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DefaultCompletionSignal != "" {
		t.Errorf("expected zero value, got %q", cfg.DefaultCompletionSignal)
	}
}

func TestLoadConfig_full(t *testing.T) {
	dir := t.TempDir()
	yaml := `
default_completion_signal: "DONE"
hooks:
  on_worktree_ready:
    - command: "echo ready"
      timeout_ms: 5000
notifications:
  on_complete: "echo done {{branch}}"
  on_error: "echo error {{branch}}"
`
	if err := os.WriteFile(filepath.Join(dir, ".amp.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DefaultCompletionSignal != "DONE" {
		t.Errorf("signal: got %q, want %q", cfg.DefaultCompletionSignal, "DONE")
	}
	if len(cfg.Hooks.OnWorktreeReady) != 1 {
		t.Fatalf("hooks: got %d, want 1", len(cfg.Hooks.OnWorktreeReady))
	}
	if cfg.Hooks.OnWorktreeReady[0].Command != "echo ready" {
		t.Errorf("hook command: got %q", cfg.Hooks.OnWorktreeReady[0].Command)
	}
	if cfg.Hooks.OnWorktreeReady[0].TimeoutMs != 5000 {
		t.Errorf("hook timeout: got %d", cfg.Hooks.OnWorktreeReady[0].TimeoutMs)
	}
	if cfg.Notifications.OnComplete != "echo done {{branch}}" {
		t.Errorf("on_complete: got %q", cfg.Notifications.OnComplete)
	}
}
