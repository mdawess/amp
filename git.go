package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

func gitRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not inside a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

// findWorktreePath returns the filesystem path of the worktree checked out on branch.
func findWorktreePath(branch string) (string, error) {
	out, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
	if err != nil {
		return "", fmt.Errorf("listing worktrees: %w", err)
	}

	var currentPath string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "worktree "):
			currentPath = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "branch "):
			b := strings.TrimPrefix(line, "branch ")
			b = strings.TrimPrefix(b, "refs/heads/")
			if b == branch {
				return currentPath, nil
			}
		}
	}

	return "", fmt.Errorf("no worktree found for branch '%s'", branch)
}
