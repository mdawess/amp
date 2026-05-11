package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSaveAndLoadRun(t *testing.T) {
	dir := t.TempDir()
	state := RunState{
		Branch:      "feature/my-branch",
		Status:      RunStatusRunning,
		StartTime:   time.Now().Truncate(time.Second),
		Prompt:      "do the thing",
		TmuxSession: "my-branch",
		LogPath:     "/tmp/amp.log",
	}
	if err := saveRun(dir, state); err != nil {
		t.Fatalf("saveRun: %v", err)
	}
	got, err := loadRun(dir, "feature/my-branch")
	if err != nil {
		t.Fatalf("loadRun: %v", err)
	}
	if got.Branch != state.Branch {
		t.Errorf("branch: got %q, want %q", got.Branch, state.Branch)
	}
	if got.Status != state.Status {
		t.Errorf("status: got %q, want %q", got.Status, state.Status)
	}
	if got.Prompt != state.Prompt {
		t.Errorf("prompt: got %q", got.Prompt)
	}
}

func TestSanitizeBranch(t *testing.T) {
	cases := []struct{ in, want string }{
		{"feature/my-branch", "feature-my-branch"},
		{"a\\b", "a-b"},
		{"a:b", "a-b"},
		{"plain", "plain"},
	}
	for _, c := range cases {
		got := sanitizeBranch(c.in)
		if got != c.want {
			t.Errorf("sanitizeBranch(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestListRuns_empty(t *testing.T) {
	dir := t.TempDir()
	runs, err := listRuns(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

func TestListRuns_multiple(t *testing.T) {
	dir := t.TempDir()
	for _, branch := range []string{"feat-a", "feat-b"} {
		s := RunState{Branch: branch, Status: RunStatusComplete, StartTime: time.Now()}
		if err := saveRun(dir, s); err != nil {
			t.Fatal(err)
		}
	}
	runs, err := listRuns(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs, got %d", len(runs))
	}
}

func TestListRuns_skipsInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	runsPath := runsDir(dir)
	if err := os.MkdirAll(runsPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runsPath, "bad.json"), []byte("{invalid}"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := RunState{Branch: "good", Status: RunStatusRunning, StartTime: time.Now()}
	if err := saveRun(dir, s); err != nil {
		t.Fatal(err)
	}
	runs, err := listRuns(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("expected 1 run, got %d", len(runs))
	}
}

func TestEnsureAmpDirGitignored(t *testing.T) {
	dir := t.TempDir()
	if err := ensureAmpDirGitignored(dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), ".amp/") {
		t.Error(".gitignore does not contain .amp/")
	}
	// Idempotent: second call should not duplicate
	if err := ensureAmpDirGitignored(dir); err != nil {
		t.Fatal(err)
	}
	data2, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	count := strings.Count(string(data2), ".amp/")
	if count != 1 {
		t.Errorf("expected .amp/ to appear once, got %d", count)
	}
}
