package stack

import "github.com/mdawes/amp/internal/shell"

// StackCmd wraps `gh stack` for managing stacked PRs.
// Requires the gh-stack extension: https://github.github.com/gh-stack/
type StackCmd struct {
	Init   StackInitCmd   `cmd:"" help:"Start a new stack (gh stack init <branch>)"`
	Add    StackAddCmd    `cmd:"" help:"Add a new layer to the stack (gh stack add <branch>)"`
	Push   StackPushCmd   `cmd:"" help:"Push all stack branches to remote (gh stack push)"`
	Submit StackSubmitCmd `cmd:"" help:"Open PRs for the stack (gh stack submit)"`
}

type StackInitCmd struct {
	Branch string   `arg:"" help:"Name of the first branch in the stack"`
	Args   []string `arg:"" optional:"" passthrough:""`
}

func (c *StackInitCmd) Run() error {
	return shell.RunPassthrough("gh", append([]string{"stack", "init", c.Branch}, c.Args...)...)
}

type StackAddCmd struct {
	Branch string   `arg:"" help:"Name of the new branch to add as the next layer"`
	Args   []string `arg:"" optional:"" passthrough:""`
}

func (c *StackAddCmd) Run() error {
	return shell.RunPassthrough("gh", append([]string{"stack", "add", c.Branch}, c.Args...)...)
}

type StackPushCmd struct {
	Args []string `arg:"" optional:"" passthrough:""`
}

func (c *StackPushCmd) Run() error {
	return shell.RunPassthrough("gh", append([]string{"stack", "push"}, c.Args...)...)
}

type StackSubmitCmd struct {
	Args []string `arg:"" optional:"" passthrough:""`
}

func (c *StackSubmitCmd) Run() error {
	return shell.RunPassthrough("gh", append([]string{"stack", "submit"}, c.Args...)...)
}
