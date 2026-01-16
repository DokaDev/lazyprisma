package commands

import (
	"bufio"
	"bytes"
	"os/exec"
	"time"
)

// CommandRunner interface for executing commands
type CommandRunner interface {
	Run(cmd *Command) error
	RunWithOutput(cmd *Command) (*CommandResult, error)
	RunAsync(cmd *Command) error
	RunAndStream(cmd *Command) error
}

type commandRunner struct {
	platform *Platform
}

// Run executes the command and waits for completion
func (r *commandRunner) Run(cmd *Command) error {
	startTime := time.Now()
	err := cmd.cmd.Run()
	duration := time.Since(startTime)

	if err != nil {
		if cmd.onError != nil {
			cmd.onError(err)
		}
		return err
	}

	exitCode := 0
	if cmd.cmd.ProcessState != nil {
		exitCode = cmd.cmd.ProcessState.ExitCode()
	}

	if cmd.onComplete != nil {
		cmd.onComplete(exitCode)
	}

	_ = duration // Track duration for potential future use

	return nil
}

// RunWithOutput executes and captures stdout/stderr
func (r *commandRunner) RunWithOutput(cmd *Command) (*CommandResult, error) {
	var stdoutBuf, stderrBuf bytes.Buffer

	cmd.cmd.Stdout = &stdoutBuf
	cmd.cmd.Stderr = &stderrBuf

	startTime := time.Now()
	err := cmd.cmd.Run()
	duration := time.Since(startTime)

	result := &CommandResult{
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		Duration: duration,
	}

	if err != nil {
		result.Error = err
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}

		if cmd.onError != nil {
			cmd.onError(err)
		}
		return result, err
	}

	result.ExitCode = 0
	if cmd.onComplete != nil {
		cmd.onComplete(0)
	}

	return result, nil
}

// RunAsync executes the command in a goroutine
func (r *commandRunner) RunAsync(cmd *Command) error {
	go func() {
		if cmd.streamOutput {
			r.RunAndStream(cmd)
		} else {
			r.RunWithOutput(cmd)
		}
	}()

	return nil
}

// RunAndStream executes with real-time output streaming
func (r *commandRunner) RunAndStream(cmd *Command) error {
	// Create pipes for stdout and stderr
	stdoutPipe, err := cmd.cmd.StdoutPipe()
	if err != nil {
		if cmd.onError != nil {
			cmd.onError(err)
		}
		return err
	}

	stderrPipe, err := cmd.cmd.StderrPipe()
	if err != nil {
		if cmd.onError != nil {
			cmd.onError(err)
		}
		return err
	}

	// Start the command
	if err := cmd.cmd.Start(); err != nil {
		if cmd.onError != nil {
			cmd.onError(err)
		}
		return err
	}

	// Create channels to signal completion of output reading
	stdoutDone := make(chan struct{})
	stderrDone := make(chan struct{})

	// Stream stdout
	go func() {
		defer close(stdoutDone)
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			if cmd.onStdout != nil {
				cmd.onStdout(line)
			}
		}
	}()

	// Stream stderr
	go func() {
		defer close(stderrDone)
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			if cmd.onStderr != nil {
				cmd.onStderr(line)
			}
		}
	}()

	// Wait for output streaming to complete
	<-stdoutDone
	<-stderrDone

	// Wait for command to finish
	err = cmd.cmd.Wait()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}

		if cmd.onError != nil {
			cmd.onError(err)
		}
	}

	if cmd.onComplete != nil {
		cmd.onComplete(exitCode)
	}

	return err
}
