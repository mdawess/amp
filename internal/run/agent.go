package run

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mdawes/amp/internal/shell"
)

const AgentPollInterval = 500 * time.Millisecond
const ExitSentinel = "__AMP_EXIT__:"

type AgentConfig struct {
	WorktreeDir      string
	Branch           string
	Prompt           string
	TmuxSession      string
	LogPath          string
	CompletionSignal string
}

// SpawnAgent creates a detached tmux session running claude --print, appending
// output to LogPath. An exit-code sentinel is written when the command finishes.
func SpawnAgent(cfg AgentConfig) error {
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
		ExitSentinel,
	)

	return shell.Run("tmux", "new-session", "-d",
		"-s", cfg.TmuxSession,
		"-c", cfg.WorktreeDir,
		"--", "sh", "-c", shellCmd,
	)
}

// WaitForCompletion polls LogPath every 500ms until the agent exits or ctx is cancelled.
func WaitForCompletion(ctx context.Context, logPath, completionSignal string) (RunStatus, int, error) {
	ticker := time.NewTicker(AgentPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return RunStatusError, -1, ctx.Err()
		case <-ticker.C:
			status, code, done, err := CheckLog(logPath, completionSignal)
			if err != nil || !done {
				continue
			}
			return status, code, nil
		}
	}
}

// CheckLog scans logPath for an exit sentinel or custom completion signal.
// Returns done=false if neither is found or the file does not yet exist.
func CheckLog(logPath, completionSignal string) (status RunStatus, exitCode int, done bool, err error) {
	data, readErr := os.ReadFile(logPath)
	if readErr != nil {
		return "", 0, false, readErr
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, ExitSentinel) {
			code, _ := strconv.Atoi(strings.TrimPrefix(line, ExitSentinel))
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

// SummarizeLog returns the last few non-empty lines of a log file as a summary.
func SummarizeLog(logPath string) string {
	data, err := os.ReadFile(logPath)
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var tail []string
	for i := len(lines) - 1; i >= 0 && len(tail) < 3; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" && !strings.HasPrefix(line, ExitSentinel) {
			tail = append([]string{line}, tail...)
		}
	}
	return strings.Join(tail, " | ")
}

func singleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
