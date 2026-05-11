package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"
)

type RunCmd struct {
	Start StartRunCmd  `cmd:"" help:"Start an autonomous agent run on a branch"`
	Ls    ListRunsCmd  `cmd:"" help:"List all runs"`
	Logs  RunLogsCmd   `cmd:"" help:"Tail the log for a run"`
}

type StartRunCmd struct {
	Branch  string `arg:"" help:"Branch to run the agent on (worktree created if absent)"`
	Prompt  string `arg:"" help:"Prompt to pass to the agent"`
	Signal  string `short:"s" help:"Completion signal string (overrides .amp.yaml default)"`
	Session string `help:"Tmux session name (default: last segment of branch)"`
	LogPath string `help:"Log file path (default: .amp/runs/<branch>.log)"`
}

func (c *StartRunCmd) Run() error {
	root, err := gitRoot()
	if err != nil {
		return err
	}

	cfg, err := loadConfig(root)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := ensureAmpDirGitignored(root); err != nil {
		return fmt.Errorf("updating .gitignore: %w", err)
	}

	worktreePath, err := ensureWorktree(root, c.Branch)
	if err != nil {
		return fmt.Errorf("ensuring worktree: %w", err)
	}

	ctx := context.Background()
	if len(cfg.Hooks.OnWorktreeReady) > 0 {
		fmt.Println("Running on_worktree_ready hooks...")
		results := runHooks(ctx, worktreePath, cfg.Hooks.OnWorktreeReady)
		for _, r := range results {
			if r.Err != nil {
				return fmt.Errorf("hook %q failed (exit %d): %s", r.Command, r.ExitCode, r.Stderr)
			}
		}
	}

	sessionName := c.Session
	if sessionName == "" {
		parts := strings.Split(c.Branch, "/")
		sessionName = parts[len(parts)-1]
	}

	logPath := c.LogPath
	if logPath == "" {
		logPath = filepath.Join(root, ".amp", "runs", sanitizeBranch(c.Branch)+".log")
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return err
	}

	completionSignal := c.Signal
	if completionSignal == "" {
		completionSignal = cfg.DefaultCompletionSignal
	}

	state := RunState{
		Branch:           c.Branch,
		Status:           RunStatusRunning,
		StartTime:        time.Now(),
		Prompt:           c.Prompt,
		TmuxSession:      sessionName,
		LogPath:          logPath,
		CompletionSignal: completionSignal,
	}
	if err := saveRun(root, state); err != nil {
		return fmt.Errorf("saving run state: %w", err)
	}

	agentCfg := AgentConfig{
		WorktreeDir:      worktreePath,
		Branch:           c.Branch,
		Prompt:           c.Prompt,
		TmuxSession:      sessionName,
		LogPath:          logPath,
		CompletionSignal: completionSignal,
	}
	if err := spawnAgent(agentCfg); err != nil {
		return fmt.Errorf("spawning agent: %w", err)
	}

	fmt.Printf("Agent started: branch=%s session=%s log=%s\n", c.Branch, sessionName, logPath)
	fmt.Println("Waiting for completion...")

	finalStatus, exitCode, waitErr := waitForCompletion(ctx, logPath, completionSignal)

	state.Status = finalStatus
	state.EndTime = time.Now()
	state.ExitCode = exitCode
	_ = saveRun(root, state)

	payload := NotifyPayload{
		Branch:  c.Branch,
		Status:  finalStatus,
		Summary: summarizeLog(logPath),
	}
	if finalStatus == RunStatusComplete {
		sendNotification(ctx, cfg.Notifications.OnComplete, payload)
	} else {
		sendNotification(ctx, cfg.Notifications.OnError, payload)
	}

	if waitErr != nil {
		return fmt.Errorf("agent exited with error: %w", waitErr)
	}
	if exitCode != 0 {
		return fmt.Errorf("agent exited with code %d", exitCode)
	}

	fmt.Printf("Run complete: branch=%s status=%s\n", c.Branch, finalStatus)
	return nil
}

type ListRunsCmd struct{}

func (c *ListRunsCmd) Run() error {
	root, err := gitRoot()
	if err != nil {
		return err
	}
	runs, err := listRuns(root)
	if err != nil {
		return err
	}
	if len(runs) == 0 {
		fmt.Println("No runs found.")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "BRANCH\tSTATUS\tSTARTED\tPROMPT")
	for _, r := range runs {
		prompt := r.Prompt
		if len(prompt) > 60 {
			prompt = prompt[:57] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			r.Branch,
			r.Status,
			r.StartTime.Format("2006-01-02 15:04"),
			prompt,
		)
	}
	return w.Flush()
}

type RunLogsCmd struct {
	Branch string `arg:"" help:"Branch name to tail logs for"`
}

func (c *RunLogsCmd) Run() error {
	root, err := gitRoot()
	if err != nil {
		return err
	}
	state, err := loadRun(root, c.Branch)
	if err != nil {
		return fmt.Errorf("no run found for branch %q: %w", c.Branch, err)
	}
	return exec.Command("tail", "-f", state.LogPath).Run()
}

// ensureWorktree returns the path to an existing worktree for branch, creating one if absent.
func ensureWorktree(repoRoot, branch string) (string, error) {
	path, err := findWorktreePath(branch)
	if err == nil {
		return path, nil
	}
	branchDir := strings.ReplaceAll(branch, "/", "-")
	worktreePath := filepath.Join(filepath.Dir(repoRoot), branchDir)
	cmd := exec.Command("git", "worktree", "add", "-B", branch, worktreePath)
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git worktree add: %s", strings.TrimSpace(string(out)))
	}
	return worktreePath, nil
}

// summarizeLog returns the last few non-empty lines of a log file as a summary.
func summarizeLog(logPath string) string {
	data, err := os.ReadFile(logPath)
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var tail []string
	for i := len(lines) - 1; i >= 0 && len(tail) < 3; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" && !strings.HasPrefix(line, exitSentinel) {
			tail = append([]string{line}, tail...)
		}
	}
	return strings.Join(tail, " | ")
}
