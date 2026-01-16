package commands

import (
	"context"
	"fmt"
	"time"
)

// RunTestSuite executes all command system tests
func RunTestSuite() {
	fmt.Println("=== LazyPrisma Command Executor Test ===\n")

	// Create a context with cancellation support
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create command builder
	builder := NewCommandBuilder(NewPlatform())

	TestGitStatus(ctx, builder)
	TestEchoCommand(builder)
	TestGitLogWithStreamHandler(builder)
	TestPingStreaming(ctx, builder)

	fmt.Println("=== All Tests Completed ===")
}

// TestGitStatus tests git status with streaming output
func TestGitStatus(ctx context.Context, builder *CommandBuilder) {
	fmt.Println("Test 1: Git Status (Streaming)")
	fmt.Println("--------------------------------")

	cmd := builder.NewWithContext(ctx, "git", "status", "--short").
		StreamOutput().
		OnStdout(func(line string) {
			fmt.Printf("[STDOUT] %s\n", line)
		}).
		OnStderr(func(line string) {
			fmt.Printf("[STDERR] %s\n", line)
		}).
		OnComplete(func(exitCode int) {
			fmt.Printf("\n✓ Command completed with exit code: %d\n\n", exitCode)
		}).
		OnError(func(err error) {
			fmt.Printf("\n✗ Command error: %v\n\n", err)
		})

	if err := cmd.RunAsync(); err != nil {
		fmt.Printf("Failed to start command: %v\n", err)
		return
	}

	time.Sleep(2 * time.Second)
}

// TestEchoCommand tests echo with captured output
func TestEchoCommand(builder *CommandBuilder) {
	fmt.Println("Test 2: Echo Command (Captured Output)")
	fmt.Println("---------------------------------------")

	echoCmd := builder.New("echo", "Hello from LazyPrisma!")
	result, err := echoCmd.RunWithOutput()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Output: %s", result.Stdout)
		fmt.Printf("Exit Code: %d\n", result.ExitCode)
		fmt.Printf("Duration: %v\n\n", result.Duration)
	}
}

// TestGitLogWithStreamHandler tests git log with StreamHandler
func TestGitLogWithStreamHandler(builder *CommandBuilder) {
	fmt.Println("Test 3: Git Log with StreamHandler")
	fmt.Println("-----------------------------------")

	logHandler := NewStreamHandler(func(stdout, stderr []string) {
		fmt.Printf("[Handler Update] %d stdout lines, %d stderr lines\n",
			len(stdout), len(stderr))
	})

	gitLogCmd := builder.New("git", "log", "--oneline", "-n", "5").
		StreamOutput().
		OnStdout(logHandler.HandleStdout).
		OnStderr(logHandler.HandleStderr).
		OnComplete(func(exitCode int) {
			stdout, stderr := logHandler.GetOutput()
			fmt.Printf("\nFinal buffered output:\n")
			for _, line := range stdout {
				fmt.Printf("  %s\n", line)
			}
			if len(stderr) > 0 {
				fmt.Printf("\nErrors:\n")
				for _, line := range stderr {
					fmt.Printf("  %s\n", line)
				}
			}
			fmt.Printf("\n✓ Git log completed (exit code: %d)\n\n", exitCode)
		})

	if err := gitLogCmd.RunAsync(); err != nil {
		fmt.Printf("Failed to start git log: %v\n", err)
		return
	}

	time.Sleep(2 * time.Second)
}

// TestPingStreaming tests ping with real-time streaming
func TestPingStreaming(ctx context.Context, builder *CommandBuilder) {
	fmt.Println("Test 4: Ping google.com (Streaming)")
	fmt.Println("------------------------------------")

	pingCmd := builder.NewWithContext(ctx, "ping", "-c", "4", "google.com").
		StreamOutput().
		OnStdout(func(line string) {
			fmt.Printf("[PING] %s\n", line)
		}).
		OnStderr(func(line string) {
			fmt.Printf("[PING ERR] %s\n", line)
		}).
		OnComplete(func(exitCode int) {
			fmt.Printf("\n✓ Ping completed with exit code: %d\n\n", exitCode)
		}).
		OnError(func(err error) {
			fmt.Printf("\n✗ Ping error: %v\n\n", err)
		})

	if err := pingCmd.RunAsync(); err != nil {
		fmt.Printf("Failed to start ping: %v\n", err)
		return
	}

	time.Sleep(5 * time.Second)
}
