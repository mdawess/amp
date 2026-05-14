# amp

CLI for managing the end-to-end agent coding workflow — worktrees, tmux windows, and stacked PRs.

## Prerequisites

- [git](https://git-scm.com/)
- [tmux](https://github.com/tmux/tmux)
- [gh CLI](https://cli.github.com/) with the [gh-stack extension](https://github.github.com/gh-stack/)

## Installation

```sh
curl -fsSL https://raw.githubusercontent.com/mdawess/amp/main/install.sh | sh
```

Installs the latest release binary to `/usr/local/bin`. Supports macOS and Linux on arm64 and amd64.

To update to the latest version:

```sh
amp update
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

### Autonomous agent runs

```sh
# Start an agent on a branch (creates the worktree if it doesn't exist)
amp run start <branch> <prompt> [--signal <completion-signal>] [--session <name>] [--log <path>]

# List all runs and their status
amp run ls

# Tail the log for a run
amp run logs <branch>
```

`amp run start` creates the worktree, runs any `on_worktree_ready` hooks, then spawns `claude --print` in a **detached** tmux session and blocks until the agent exits. A push notification is fired on completion via the commands in `.amp.yaml`.

**Keeping your machine awake** — wrap with `caffeinate -d` to prevent sleep while the agent runs:

```sh
caffeinate -d amp run start my-feature "implement the auth module per the PRD"
```

**Monitoring over SSH** — because the agent runs in a detached tmux session you can attach to it from any machine on your network without interrupting the run:

```sh
# On your phone or another machine
ssh user@your-machine
tmux attach -t <session-name>   # session name = last segment of branch by default
# Detach without killing: Ctrl+B D
```

The `amp run start` process polling in the original terminal is unaffected by attach/detach.

#### Configuration — `.amp.yaml`

Place `.amp.yaml` in the repo root to configure hooks and notifications:

```yaml
# Shell commands run inside the worktree before the agent starts.
# Useful for installing dependencies, running migrations, etc.
hooks:
  on_worktree_ready:
    - command: "npm install"
      timeout_ms: 60000
    - command: "cp .env.example .env"

# Shell commands fired when a run finishes.
# Supports {{branch}}, {{status}}, and {{summary}} placeholders.
notifications:
  on_complete: "osascript -e 'display notification \"{{summary}}\" with title \"amp: {{branch}} done\"'"
  on_error:    "osascript -e 'display notification \"{{summary}}\" with title \"amp: {{branch}} failed\"'"

# A string in the agent's output that signals successful completion
# (in addition to a zero exit code).
default_completion_signal: ""
```

Run state is persisted to `.amp/runs/<branch>.json` (automatically gitignored).

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
