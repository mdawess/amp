package run

import (
	"context"
	"testing"
	"time"
)

func TestRunHook_success(t *testing.T) {
	result := RunHook(context.Background(), t.TempDir(), Hook{Command: "echo hello"})
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.ExitCode != 0 {
		t.Errorf("exit code: got %d, want 0", result.ExitCode)
	}
	if result.Stdout != "hello\n" {
		t.Errorf("stdout: got %q, want %q", result.Stdout, "hello\n")
	}
}

func TestRunHook_failure(t *testing.T) {
	result := RunHook(context.Background(), t.TempDir(), Hook{Command: "exit 42"})
	if result.ExitCode != 42 {
		t.Errorf("exit code: got %d, want 42", result.ExitCode)
	}
	if result.Err == nil {
		t.Error("expected non-nil error")
	}
}

func TestRunHook_timeout(t *testing.T) {
	result := RunHook(context.Background(), t.TempDir(), Hook{
		Command:   "sleep 10",
		TimeoutMs: 50,
	})
	if !result.TimedOut {
		t.Error("expected TimedOut = true")
	}
}

func TestRunHooks_order(t *testing.T) {
	dir := t.TempDir()
	hooks := []Hook{
		{Command: "echo first"},
		{Command: "echo second"},
	}
	results := RunHooks(context.Background(), dir, hooks)
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Stdout != "first\n" {
		t.Errorf("result[0]: got %q", results[0].Stdout)
	}
	if results[1].Stdout != "second\n" {
		t.Errorf("result[1]: got %q", results[1].Stdout)
	}
}

func TestRunHook_cancelledContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	result := RunHook(ctx, t.TempDir(), Hook{Command: "sleep 10"})
	if result.Err == nil {
		t.Error("expected error from cancelled context")
	}
}
