package run

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/mdawes/amp/internal/git"
	"github.com/mdawes/amp/internal/shell"
)

type RunCmd struct {
	Start StartRunCmd `cmd:"" help:"Start an autonomous agent run on a branch"`
	Ls    ListRunsCmd `cmd:"" help:"List all runs"`
	Logs  RunLogsCmd  `cmd:"" help:"Tail the log for a run"`
}

type StartRunCmd struct {
	Branch  string `arg:"" help:"Branch to run the agent on (worktree created if absent)"`
	Prompt  string `arg:"" help:"Prompt to pass to the agent"`
	Signal  string `short:"s" help:"Completion signal string (overrides .amp.yaml default)"`
	Session string `help:"Tmux session name (default: last segment of branch)"`
	LogPath string `help:"Log file path (default: .amp/runs/<branch>.log)"`
}

func (c *StartRunCmd) Run() error {
	root, err := git.Root()
	if err != nil {
		return err
	}

	cfg, err := LoadConfig(root)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := EnsureAmpDirGitignored(root); err != nil {
		return fmt.Errorf("updating .gitignore: %w", err)
	}

	worktreePath, err := ensureWorktree(root, c.Branch)
	if err != nil {
		return fmt.Errorf("ensuring worktree: %w", err)
	}

	ctx := context.Background()
	if len(cfg.Hooks.OnWorktreeReady) > 0 {
		fmt.Println("Running on_worktree_ready hooks...")
		results := RunHooks(ctx, worktreePath, cfg.Hooks.OnWorktreeReady)
		for _, r := range results {
			if r.Err != nil {
				return fmt.Errorf("hook %q failed (exit %d): %s", r.Command, r.ExitCode, r.Stderr)
			}
		}
	}

	sessionName := c.Session
	if sessionName == "" {
		sessionName = git.DefaultSessionName(c.Branch)
	}

	logPath := c.LogPath
	if logPath == "" {
		logPath = filepath.Join(root, ".amp", "runs", SanitizeBranch(c.Branch)+".log")
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
	if err := SaveRun(root, state); err != nil {
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
	if err := SpawnAgent(agentCfg); err != nil {
		return fmt.Errorf("spawning agent: %w", err)
	}

	fmt.Printf("Agent started: branch=%s session=%s log=%s\n", c.Branch, sessionName, logPath)
	fmt.Println("Waiting for completion...")

	finalStatus, exitCode, waitErr := WaitForCompletion(ctx, logPath, completionSignal)

	state.Status = finalStatus
	state.EndTime = time.Now()
	state.ExitCode = exitCode
	_ = SaveRun(root, state)

	payload := NotifyPayload{
		Branch:  c.Branch,
		Status:  finalStatus,
		Summary: SummarizeLog(logPath),
	}
	if finalStatus == RunStatusComplete {
		SendNotification(ctx, cfg.Notifications.OnComplete, payload)
	} else {
		SendNotification(ctx, cfg.Notifications.OnError, payload)
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
	root, err := git.Root()
	if err != nil {
		return err
	}
	runs, err := ListRuns(root)
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
	root, err := git.Root()
	if err != nil {
		return err
	}
	state, err := LoadRun(root, c.Branch)
	if err != nil {
		return fmt.Errorf("no run found for branch %q: %w", c.Branch, err)
	}
	return exec.Command("tail", "-f", state.LogPath).Run()
}

func ensureWorktree(repoRoot, branch string) (string, error) {
	path, err := git.FindWorktreePath(branch)
	if err == nil {
		return path, nil
	}
	worktreePath := git.WorktreeSiblingPath(repoRoot, branch)
	if err := shell.Run("git", "worktree", "add", "-B", branch, worktreePath); err != nil {
		return "", fmt.Errorf("creating worktree: %w", err)
	}
	return worktreePath, nil
}
