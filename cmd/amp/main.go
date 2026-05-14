package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/mdawes/amp/internal/run"
	"github.com/mdawes/amp/internal/stack"
	"github.com/mdawes/amp/internal/worktree"
)

var cli struct {
	Worktree worktree.WorktreeCmd `cmd:"" help:"Manage git worktrees and tmux sessions"`
	Stack    stack.StackCmd       `cmd:"" help:"Manage stacked PRs via gh stack"`
	Window   worktree.WindowCmd   `cmd:"" help:"Open a new tmux window in the current session"`
	Run      run.RunCmd           `cmd:"" help:"Run an autonomous agent on a branch"`
	Update   UpdateCmd            `cmd:"" help:"Update amp to the latest release"`
}

func main() {
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "--help")
	}
	ctx := kong.Parse(&cli,
		kong.Name("amp"),
		kong.Description("CLI for managing the end-to-end agent coding workflow"),
		kong.UsageOnError(),
	)
	ctx.FatalIfErrorf(ctx.Run())
}
