# amp

CLI for managing the end-to-end agent coding workflow — worktrees, tmux windows, and stacked PRs.

## Prerequisites

- [git](https://git-scm.com/)
- [tmux](https://github.com/tmux/tmux)
- [gh CLI](https://cli.github.com/) with the [gh-stack extension](https://github.github.com/gh-stack/)

## Installation

```sh
make install
```

Installs the binary to `$(go env GOPATH)/bin` (typically `~/go/bin`). Make sure that directory is on your `PATH`.

To update to the latest version:

```sh
make update
```

## Commands

### Worktrees

```sh
# Create a worktree for a new branch and open it in a tmux session
amp worktree new <branch> [-s session-name] [-p path]

# Open an existing worktree in a new tmux session
amp worktree open <branch> [-s session-name]

# Remove a worktree and prune git refs
amp worktree rm <branch> [-f]

# List all worktrees
amp worktree ls
```

`worktree new` creates the worktree as a sibling directory to the repo root (e.g. `../my-feature`) and opens a new tmux session pointed at it. The session name defaults to the last segment of the branch name.

### Windows

```sh
# Open a new tmux window in the current session
amp window <name> [-c dir]
```

### Stacked PRs

```sh
amp stack init <branch>    # start a new stack
amp stack add <branch>     # add the next layer
amp stack push             # push all branches to remote
amp stack submit           # open PRs for the stack
```

**Typical workflow:**

```sh
amp worktree new feature/auth        # branch + tmux window
amp stack init auth-layer            # start the stack

# write code, commit

amp stack add api-routes             # next layer
# write code, commit

amp stack push                       # push all branches
amp stack submit                     # open PRs

amp worktree rm feature/auth         # clean up when done
```

All `stack` subcommands pass extra arguments through to `gh stack`.
