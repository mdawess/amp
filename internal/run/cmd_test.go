package run

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEnsureWorktree_error(t *testing.T) {
	_, err := ensureWorktree("/nonexistent/root", "no-such-branch")
	if err == nil {
		t.Error("expected error for nonexistent repo")
	}
}

func TestSummarizeLog_empty(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "empty.log")
	os.WriteFile(logPath, []byte(""), 0o644)
	got := SummarizeLog(logPath)
	if got != "" {
		t.Errorf("expected empty summary, got %q", got)
	}
}

func TestSummarizeLog_skipsExitSentinel(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "run.log")
	content := "line one\nline two\nline three\n" + ExitSentinel + "0\n"
	os.WriteFile(logPath, []byte(content), 0o644)
	got := SummarizeLog(logPath)
	if strings.Contains(got, ExitSentinel) {
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
	got := SummarizeLog(logPath)
	parts := strings.Split(got, " | ")
	if len(parts) > 3 {
		t.Errorf("summary should have at most 3 parts, got %d: %q", len(parts), got)
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
	if err := SaveRun(dir, state); err != nil {
		t.Fatal(err)
	}
	got, err := LoadRun(dir, "feat/x")
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
