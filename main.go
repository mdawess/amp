package main

import (
	"os"

	"github.com/alecthomas/kong"
)

var cli struct {
	Worktree WorktreeCmd `cmd:"" help:"Manage git worktrees and tmux sessions"`
	Stack    StackCmd    `cmd:"" help:"Manage stacked PRs via gh stack"`
	Window   WindowCmd   `cmd:"" help:"Open a new tmux window in the current session"`
	Run      RunCmd      `cmd:"" help:"Run an autonomous agent on a branch"`
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
