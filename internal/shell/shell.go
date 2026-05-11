package shell

import (
	"os"
	"os/exec"
)

// Run executes a command, streaming stdout/stderr to the terminal.
func Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunPassthrough executes a command with stdin/stdout/stderr all attached,
// suitable for interactive commands.
func RunPassthrough(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
