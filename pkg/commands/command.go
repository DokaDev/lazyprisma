package commands

import (
	"context"
	"os/exec"
	"syscall"
	"time"
)

// Command represents a shell command to be executed
type Command struct {
	cmd    *exec.Cmd
	ctx    context.Context
	runner CommandRunner

	// Configuration options
	streamOutput bool   // Stream output in real-time
	captureOutput bool   // Capture output for return
	workingDir   string // Working directory
	envVars      []string // Additional environment variables

	// Callbacks for async execution
	onStdout   func(string) // Called for each stdout line
	onStderr   func(string) // Called for each stderr line
	onComplete func(int)    // Called on completion with exit code
	onError    func(error)  // Called on error
}

// CommandResult holds the result of command execution
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
	Duration time.Duration
}

// WithWorkingDir sets the working directory
func (c *Command) WithWorkingDir(dir string) *Command {
	c.workingDir = dir
	c.cmd.Dir = dir
	return c
}

// WithEnv adds environment variables
func (c *Command) WithEnv(vars ...string) *Command {
	c.envVars = append(c.envVars, vars...)
	c.cmd.Env = append(c.cmd.Env, vars...)
	return c
}

// StreamOutput enables real-time output streaming
func (c *Command) StreamOutput() *Command {
	c.streamOutput = true
	return c
}

// CaptureOutput enables output capture for return
func (c *Command) CaptureOutput() *Command {
	c.captureOutput = true
	return c
}

// OnStdout sets callback for stdout lines
func (c *Command) OnStdout(fn func(string)) *Command {
	c.onStdout = fn
	return c
}

// OnStderr sets callback for stderr lines
func (c *Command) OnStderr(fn func(string)) *Command {
	c.onStderr = fn
	return c
}

// OnComplete sets callback for completion
func (c *Command) OnComplete(fn func(int)) *Command {
	c.onComplete = fn
	return c
}

// OnError sets callback for errors
func (c *Command) OnError(fn func(error)) *Command {
	c.onError = fn
	return c
}

// Run executes the command synchronously
func (c *Command) Run() error {
	return c.runner.Run(c)
}

// RunWithOutput executes and captures output
func (c *Command) RunWithOutput() (*CommandResult, error) {
	return c.runner.RunWithOutput(c)
}

// RunAsync executes the command asynchronously
func (c *Command) RunAsync() error {
	return c.runner.RunAsync(c)
}

// RunAndStream executes with real-time streaming
func (c *Command) RunAndStream() error {
	return c.runner.RunAndStream(c)
}

// Kill terminates the running process and its children (process group)
func (c *Command) Kill() error {
	if c.cmd != nil && c.cmd.Process != nil {
		// Use negative PID to kill the process group
		// Requires SysProcAttr.Setpgid to be true (set in CommandBuilder)
		return syscall.Kill(-c.cmd.Process.Pid, syscall.SIGKILL)
	}
	return nil
}
