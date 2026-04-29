package main

import (
	"os"
	"os/exec"
)

// run executes a command, streaming stdout/stderr to the terminal.
func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runPassthrough executes a command with stdin/stdout/stderr all attached,
// suitable for interactive commands.
func runPassthrough(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
