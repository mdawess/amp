package main

import (
	"os"

	"github.com/alecthomas/kong"
)

var cli struct {
	Worktree WorktreeCmd `cmd:"" help:"Manage git worktrees and tmux windows"`
	Stack    StackCmd    `cmd:"" help:"Manage stacked PRs via gh stack"`
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
