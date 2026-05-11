package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const agentPollInterval = 500 * time.Millisecond
const exitSentinel = "__AMP_EXIT__:"

type AgentConfig struct {
	WorktreeDir      string
	Branch           string
	Prompt           string
	TmuxSession      string
	LogPath          string
	CompletionSignal string
}

// spawnAgent creates a detached tmux session running claude --print, appending
// output to LogPath. An exit-code sentinel is written when the command finishes.
func spawnAgent(cfg AgentConfig) error {
	// Write prompt to temp file to avoid shell quoting issues with arbitrary content.
	f, err := os.CreateTemp("", "amp-prompt-*")
	if err != nil {
		return fmt.Errorf("creating prompt file: %w", err)
	}
	promptPath := f.Name()
	if _, err := f.WriteString(cfg.Prompt); err != nil {
		f.Close()
		os.Remove(promptPath)
		return fmt.Errorf("writing prompt file: %w", err)
	}
	f.Close()

	shellCmd := fmt.Sprintf(
		`claude --print "$(cat %[1]s)" >> %[2]s 2>&1; _code=$?; rm -f %[1]s; echo %[3]s$_code >> %[2]s`,
		singleQuote(promptPath),
		singleQuote(cfg.LogPath),
		exitSentinel,
	)

	return run("tmux", "new-session", "-d",
		"-s", cfg.TmuxSession,
		"-c", cfg.WorktreeDir,
		"--", "sh", "-c", shellCmd,
	)
}

// waitForCompletion polls LogPath every 500ms until the agent exits or ctx is cancelled.
func waitForCompletion(ctx context.Context, logPath, completionSignal string) (RunStatus, int, error) {
	ticker := time.NewTicker(agentPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return RunStatusError, -1, ctx.Err()
		case <-ticker.C:
			status, code, done, err := checkLog(logPath, completionSignal)
			if err != nil || !done {
				continue
			}
			return status, code, nil
		}
	}
}

// checkLog scans logPath for an exit sentinel or custom completion signal.
// Returns done=false if neither is found or the file does not yet exist.
func checkLog(logPath, completionSignal string) (status RunStatus, exitCode int, done bool, err error) {
	data, readErr := os.ReadFile(logPath)
	if readErr != nil {
		return "", 0, false, readErr
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, exitSentinel) {
			code, _ := strconv.Atoi(strings.TrimPrefix(line, exitSentinel))
			if code == 0 {
				return RunStatusComplete, code, true, nil
			}
			return RunStatusError, code, true, nil
		}
		if completionSignal != "" && strings.Contains(line, completionSignal) {
			return RunStatusComplete, 0, true, nil
		}
	}
	return "", 0, false, nil
}

// singleQuote wraps s in single quotes, escaping any embedded single quotes.
func singleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
