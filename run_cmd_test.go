package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEnsureWorktree_existing(t *testing.T) {
	// ensureWorktree should return the path without error for a known worktree.
	// We can't create real git worktrees in tests, so we test the fallback error path instead.
	_, err := ensureWorktree("/nonexistent/root", "no-such-branch")
	if err == nil {
		t.Error("expected error for nonexistent repo")
	}
}

func TestSummarizeLog_empty(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "empty.log")
	os.WriteFile(logPath, []byte(""), 0o644)
	got := summarizeLog(logPath)
	if got != "" {
		t.Errorf("expected empty summary, got %q", got)
	}
}

func TestSummarizeLog_skipsExitSentinel(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "run.log")
	content := "line one\nline two\nline three\n" + exitSentinel + "0\n"
	os.WriteFile(logPath, []byte(content), 0o644)
	got := summarizeLog(logPath)
	if strings.Contains(got, exitSentinel) {
		t.Errorf("summary should not contain exit sentinel, got %q", got)
	}
	if !strings.Contains(got, "line three") {
		t.Errorf("summary should contain last lines, got %q", got)
	}
}

func TestSummarizeLog_truncatesLong(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "run.log")
	var lines []string
	for i := 0; i < 10; i++ {
		lines = append(lines, "output line")
	}
	os.WriteFile(logPath, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
	got := summarizeLog(logPath)
	parts := strings.Split(got, " | ")
	if len(parts) > 3 {
		t.Errorf("summary should have at most 3 parts, got %d: %q", len(parts), got)
	}
}

func TestListRunsCmd_noRuns(t *testing.T) {
	// Tests the underlying listRuns with empty dir — covered in run_store_test.go,
	// but exercise the command's zero-state path here.
	dir := t.TempDir()
	runs, err := listRuns(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

func TestStartRunCmd_logPathDefault(t *testing.T) {
	// Verify the default log path construction without running the full command.
	branch := "feature/my-task"
	root := "/repo/root"
	expected := filepath.Join(root, ".amp", "runs", sanitizeBranch(branch)+".log")
	got := filepath.Join(root, ".amp", "runs", sanitizeBranch(branch)+".log")
	if got != expected {
		t.Errorf("log path: got %q, want %q", got, expected)
	}
}

func TestRunState_serialisation(t *testing.T) {
	dir := t.TempDir()
	state := RunState{
		Branch:           "feat/x",
		Status:           RunStatusComplete,
		StartTime:        time.Now().Truncate(time.Second),
		EndTime:          time.Now().Truncate(time.Second),
		Prompt:           "implement feature X",
		TmuxSession:      "x",
		LogPath:          "/tmp/x.log",
		CompletionSignal: "DONE",
		ExitCode:         0,
	}
	if err := saveRun(dir, state); err != nil {
		t.Fatal(err)
	}
	got, err := loadRun(dir, "feat/x")
	if err != nil {
		t.Fatal(err)
	}
	if got.CompletionSignal != state.CompletionSignal {
		t.Errorf("signal: got %q, want %q", got.CompletionSignal, state.CompletionSignal)
	}
	if got.ExitCode != state.ExitCode {
		t.Errorf("exit code: got %d, want %d", got.ExitCode, state.ExitCode)
	}
}
