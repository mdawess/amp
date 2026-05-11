package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// worktreeSiblingPath returns the default path for a worktree: a sibling directory of repoRoot.
func worktreeSiblingPath(repoRoot, branch string) string {
	return filepath.Join(filepath.Dir(repoRoot), strings.ReplaceAll(branch, "/", "-"))
}

// defaultSessionName returns the default tmux session name for a branch: its last path segment.
func defaultSessionName(branch string) string {
	parts := strings.Split(branch, "/")
	return parts[len(parts)-1]
}

type WorktreeCmd struct {
	New  NewWorktreeCmd  `cmd:"" help:"Create a new worktree and open it in a tmux session"`
	Open OpenWorktreeCmd `cmd:"" help:"Open an existing worktree in a new tmux session"`
	Rm   RmWorktreeCmd   `cmd:"" help:"Remove a worktree and prune refs"`
	Ls   ListWorktreeCmd `cmd:"" help:"List all worktrees"`
}

type NewWorktreeCmd struct {
	Branch  string `arg:"" help:"Branch name to create"`
	Session string `short:"s" help:"Tmux session name (default: last segment of branch name)"`
	Path    string `short:"p" help:"Worktree path (default: sibling directory named after branch)"`
}

func (c *NewWorktreeCmd) Run() error {
	root, err := gitRoot()
	if err != nil {
		return err
	}

	worktreePath := c.Path
	if worktreePath == "" {
		worktreePath = worktreeSiblingPath(root, c.Branch)
	}

	if err := run("git", "worktree", "add", "-B", c.Branch, worktreePath); err != nil {
		return fmt.Errorf("creating worktree: %w", err)
	}

	sessionName := c.Session
	if sessionName == "" {
		sessionName = defaultSessionName(c.Branch)
	}

	if os.Getenv("TMUX") == "" {
		fmt.Printf("Created worktree at %s (not in tmux, skipping session creation)\n", worktreePath)
		return nil
	}

	if err := run("tmux", "new-session", "-d", "-s", sessionName, "-c", worktreePath); err != nil {
		return fmt.Errorf("creating tmux session: %w", err)
	}
	if err := run("tmux", "switch-client", "-t", sessionName); err != nil {
		return fmt.Errorf("switching to tmux session: %w", err)
	}

	fmt.Printf("Created worktree at %s, opened tmux session '%s'\n", worktreePath, sessionName)
	return nil
}

type OpenWorktreeCmd struct {
	Branch  string `arg:"" help:"Branch name of the worktree to open"`
	Session string `short:"s" help:"Tmux session name (default: last segment of branch name)"`
}

func (c *OpenWorktreeCmd) Run() error {
	path, err := findWorktreePath(c.Branch)
	if err != nil {
		return err
	}

	sessionName := c.Session
	if sessionName == "" {
		sessionName = defaultSessionName(c.Branch)
	}

	if os.Getenv("TMUX") == "" {
		fmt.Printf("Worktree at %s (not in tmux, skipping session creation)\n", path)
		return nil
	}

	if err := run("tmux", "new-session", "-d", "-s", sessionName, "-c", path); err != nil {
		return fmt.Errorf("creating tmux session: %w", err)
	}
	if err := run("tmux", "switch-client", "-t", sessionName); err != nil {
		return fmt.Errorf("switching to tmux session: %w", err)
	}

	fmt.Printf("Opened worktree '%s' in tmux session '%s'\n", path, sessionName)
	return nil
}

type RmWorktreeCmd struct {
	Branch string `arg:"" help:"Branch name of the worktree to remove"`
	Force  bool   `short:"f" help:"Force removal even with uncommitted changes"`
}

func (c *RmWorktreeCmd) Run() error {
	path, err := findWorktreePath(c.Branch)
	if err != nil {
		return err
	}

	args := []string{"worktree", "remove"}
	if c.Force {
		args = append(args, "--force")
	}
	args = append(args, path)

	if err := run("git", args...); err != nil {
		return fmt.Errorf("removing worktree: %w", err)
	}

	if err := run("git", "worktree", "prune"); err != nil {
		return fmt.Errorf("pruning worktrees: %w", err)
	}

	fmt.Printf("Removed worktree for branch '%s'\n", c.Branch)
	return nil
}

type ListWorktreeCmd struct{}

func (c *ListWorktreeCmd) Run() error {
	return run("git", "worktree", "list")
}

// WindowCmd creates a new tmux window in the current session.
type WindowCmd struct {
	Name string `arg:"" help:"Window name"`
	Dir  string `short:"c" help:"Starting directory (default: current directory)"`
}

func (c *WindowCmd) Run() error {
	if os.Getenv("TMUX") == "" {
		return fmt.Errorf("not inside a tmux session")
	}

	args := []string{"new-window", "-n", c.Name}
	if c.Dir != "" {
		args = append(args, "-c", c.Dir)
	}

	if err := run("tmux", args...); err != nil {
		return fmt.Errorf("creating tmux window: %w", err)
	}

	fmt.Printf("Opened tmux window '%s'\n", c.Name)
	return nil
}
