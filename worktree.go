package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type WorktreeCmd struct {
	New NewWorktreeCmd  `cmd:"" help:"Create a new worktree and open it in a tmux window"`
	Rm  RmWorktreeCmd   `cmd:"" help:"Remove a worktree and prune refs"`
	Ls  ListWorktreeCmd `cmd:"" help:"List all worktrees"`
}

type NewWorktreeCmd struct {
	Branch string `arg:"" help:"Branch name to create"`
	Window string `short:"w" help:"Tmux window name (default: last segment of branch name)"`
	Path   string `short:"p" help:"Worktree path (default: sibling directory named after branch)"`
}

func (c *NewWorktreeCmd) Run() error {
	root, err := gitRoot()
	if err != nil {
		return err
	}

	worktreePath := c.Path
	if worktreePath == "" {
		branchDir := strings.ReplaceAll(c.Branch, "/", "-")
		worktreePath = filepath.Join(filepath.Dir(root), branchDir)
	}

	if err := run("git", "worktree", "add", "-b", c.Branch, worktreePath); err != nil {
		return fmt.Errorf("creating worktree: %w", err)
	}

	windowName := c.Window
	if windowName == "" {
		parts := strings.Split(c.Branch, "/")
		windowName = parts[len(parts)-1]
	}

	if os.Getenv("TMUX") == "" {
		fmt.Printf("Created worktree at %s (not in tmux, skipping window creation)\n", worktreePath)
		return nil
	}

	if err := run("tmux", "new-window", "-n", windowName, "-c", worktreePath); err != nil {
		return fmt.Errorf("creating tmux window: %w", err)
	}

	fmt.Printf("Created worktree at %s, opened tmux window '%s'\n", worktreePath, windowName)
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
