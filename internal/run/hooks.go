package run

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

type HookResult struct {
	Command  string
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
	TimedOut bool
}

func RunHooks(ctx context.Context, dir string, hooks []Hook) []HookResult {
	results := make([]HookResult, 0, len(hooks))
	for _, h := range hooks {
		results = append(results, RunHook(ctx, dir, h))
	}
	return results
}

func RunHook(ctx context.Context, dir string, h Hook) HookResult {
	hookCtx := ctx
	var cancel context.CancelFunc
	if h.TimeoutMs > 0 {
		hookCtx, cancel = context.WithTimeout(ctx, time.Duration(h.TimeoutMs)*time.Millisecond)
		defer cancel()
	}

	cmd := exec.CommandContext(hookCtx, "sh", "-c", h.Command)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := HookResult{
		Command: h.Command,
		Stdout:  stdout.String(),
		Stderr:  stderr.String(),
		Err:     err,
	}
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	if hookCtx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
	}
	return result
}
