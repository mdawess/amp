package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCheckLog_notFound(t *testing.T) {
	_, _, done, err := checkLog("/nonexistent/path.log", "")
	if err == nil {
		t.Error("expected error for missing file")
	}
	if done {
		t.Error("expected done=false")
	}
}

func TestCheckLog_exitZero(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "out.log")
	os.WriteFile(logPath, []byte("some output\n"+exitSentinel+"0\n"), 0o644)

	status, code, done, err := checkLog(logPath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !done {
		t.Fatal("expected done=true")
	}
	if status != RunStatusComplete {
		t.Errorf("status: got %q, want %q", status, RunStatusComplete)
	}
	if code != 0 {
		t.Errorf("code: got %d, want 0", code)
	}
}

func TestCheckLog_exitNonZero(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "out.log")
	os.WriteFile(logPath, []byte(exitSentinel+"1\n"), 0o644)

	status, code, done, _ := checkLog(logPath, "")
	if !done {
		t.Fatal("expected done=true")
	}
	if status != RunStatusError {
		t.Errorf("status: got %q, want %q", status, RunStatusError)
	}
	if code != 1 {
		t.Errorf("code: got %d, want 1", code)
	}
}

func TestCheckLog_completionSignal(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "out.log")
	os.WriteFile(logPath, []byte("running...\nTask complete! All tests pass.\n"), 0o644)

	status, _, done, _ := checkLog(logPath, "Task complete!")
	if !done {
		t.Fatal("expected done=true")
	}
	if status != RunStatusComplete {
		t.Errorf("status: got %q, want %q", status, RunStatusComplete)
	}
}

func TestCheckLog_noSignalYet(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "out.log")
	os.WriteFile(logPath, []byte("still working...\n"), 0o644)

	_, _, done, err := checkLog(logPath, "DONE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if done {
		t.Error("expected done=false")
	}
}

func TestWaitForCompletion_contextCancel(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "out.log")
	os.WriteFile(logPath, []byte("running...\n"), 0o644)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	status, _, err := waitForCompletion(ctx, logPath, "")
	if err == nil {
		t.Error("expected error from cancelled context")
	}
	if status != RunStatusError {
		t.Errorf("status: got %q, want %q", status, RunStatusError)
	}
}

func TestWaitForCompletion_detectsExit(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "out.log")
	os.WriteFile(logPath, []byte("starting\n"), 0o644)

	go func() {
		time.Sleep(150 * time.Millisecond)
		f, _ := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0o644)
		f.WriteString(exitSentinel + "0\n")
		f.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	status, code, err := waitForCompletion(ctx, logPath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != RunStatusComplete {
		t.Errorf("status: got %q, want %q", status, RunStatusComplete)
	}
	if code != 0 {
		t.Errorf("code: got %d, want 0", code)
	}
}

func TestSingleQuote(t *testing.T) {
	cases := []struct{ in, want string }{
		{"plain", "'plain'"},
		{"/path/to/file", "'/path/to/file'"},
		{"it's", `'it'\''s'`},
	}
	for _, c := range cases {
		got := singleQuote(c.in)
		if got != c.want {
			t.Errorf("singleQuote(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
