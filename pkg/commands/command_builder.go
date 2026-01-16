package commands

import (
	"context"
	"os/exec"
	"syscall"
)

// CommandBuilder provides a fluent API for building commands
type CommandBuilder struct {
	runner   CommandRunner
	platform *Platform
}

// NewCommandBuilder creates a new command builder
func NewCommandBuilder(platform *Platform) *CommandBuilder {
	return &CommandBuilder{
		runner:   &commandRunner{platform: platform},
		platform: platform,
	}
}

// New creates a command from arguments
// Example: New("npx", "prisma", "migrate", "dev")
func (b *CommandBuilder) New(args ...string) *Command {
	return b.NewWithContext(context.Background(), args...)
}

// NewWithContext creates a command with a context for cancellation
func (b *CommandBuilder) NewWithContext(ctx context.Context, args ...string) *Command {
	if len(args) == 0 {
		panic("command requires at least one argument")
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	
	// Create a new process group for process management (Kill via -PID)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return &Command{
		cmd:    cmd,
		ctx:    ctx,
		runner: b.runner,
	}
}

// NewShell creates a command from a shell string
// Example: NewShell("npx prisma migrate dev --name init")
func (b *CommandBuilder) NewShell(ctx context.Context, cmdStr string) *Command {
	shell, shellArg := b.platform.GetShell()
	cmd := exec.CommandContext(ctx, shell, shellArg, cmdStr)
	
	// Create a new process group for process management (Kill via -PID)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return &Command{
		cmd:    cmd,
		ctx:    ctx,
		runner: b.runner,
	}
}
