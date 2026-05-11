package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Root returns the root directory of the git repository.
func Root() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not inside a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

// FindWorktreePath returns the filesystem path of the worktree checked out on branch.
func FindWorktreePath(branch string) (string, error) {
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

// WorktreeSiblingPath returns the default path for a worktree: a sibling directory of repoRoot.
func WorktreeSiblingPath(repoRoot, branch string) string {
	return filepath.Join(filepath.Dir(repoRoot), strings.ReplaceAll(branch, "/", "-"))
}

// DefaultSessionName returns the default tmux session name for a branch: its last path segment.
func DefaultSessionName(branch string) string {
	parts := strings.Split(branch, "/")
	return parts[len(parts)-1]
}
