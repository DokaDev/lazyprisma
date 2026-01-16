package prisma

// GenerateOptions holds options for prisma generate command
type GenerateOptions struct {
	Schema      string // Optional path to schema file
	Watch       bool   // Enable watch mode
	NoEngine    bool   // Skip engine download
	DataProxy   bool   // Generate for Data Proxy
	Accelerate  bool   // Generate for Accelerate
}

// GenerateResult holds the result of prisma generate
type GenerateResult struct {
	Success bool   // True if generation succeeded
	Output  string // Full output from generate command
	Error   string // Error message if failed
}

// Generate runs `npx prisma generate` to generate Prisma Client
func Generate(projectDir string, opts *GenerateOptions) (*GenerateResult, error) {
	// Build command args
	args := []string{"prisma", "generate"}

	if opts != nil {
		if opts.Schema != "" {
			args = append(args, "--schema", opts.Schema)
		}
		if opts.Watch {
			args = append(args, "--watch")
		}
		if opts.NoEngine {
			args = append(args, "--no-engine")
		}
		if opts.DataProxy {
			args = append(args, "--data-proxy")
		}
		if opts.Accelerate {
			args = append(args, "--accelerate")
		}
	}

	// Execute command (prepend "npx" to args)
	cmdArgs := append([]string{"npx"}, args...)
	cmd := cmdBuilder.New(cmdArgs...).WithWorkingDir(projectDir)
	result, err := cmd.RunWithOutput()

	generateResult := &GenerateResult{
		Output: result.Stdout + result.Stderr,
	}

	if err != nil || result.ExitCode != 0 {
		generateResult.Success = false
		generateResult.Error = result.Stderr
		if generateResult.Error == "" && err != nil {
			generateResult.Error = err.Error()
		}
		return generateResult, err
	}

	generateResult.Success = true
	return generateResult, nil
}

// GenerateAsync runs prisma generate asynchronously with callbacks
type GenerateCallbacks struct {
	OnStdout   func(string) // Called for each stdout line
	OnStderr   func(string) // Called for each stderr line
	OnComplete func(bool)   // Called when complete (success bool)
	OnError    func(error)  // Called on error
}

// GenerateAsync runs `npx prisma generate` asynchronously with real-time output
func GenerateAsync(projectDir string, opts *GenerateOptions, callbacks *GenerateCallbacks) error {
	// Build command args
	args := []string{"prisma", "generate"}

	if opts != nil {
		if opts.Schema != "" {
			args = append(args, "--schema", opts.Schema)
		}
		if opts.Watch {
			args = append(args, "--watch")
		}
		if opts.NoEngine {
			args = append(args, "--no-engine")
		}
		if opts.DataProxy {
			args = append(args, "--data-proxy")
		}
		if opts.Accelerate {
			args = append(args, "--accelerate")
		}
	}

	// Build command with callbacks (prepend "npx" to args)
	cmdArgs := append([]string{"npx"}, args...)
	cmd := cmdBuilder.New(cmdArgs...).
		WithWorkingDir(projectDir).
		StreamOutput()

	if callbacks != nil {
		if callbacks.OnStdout != nil {
			cmd.OnStdout(callbacks.OnStdout)
		}
		if callbacks.OnStderr != nil {
			cmd.OnStderr(callbacks.OnStderr)
		}
		if callbacks.OnComplete != nil {
			cmd.OnComplete(func(exitCode int) {
				callbacks.OnComplete(exitCode == 0)
			})
		}
		if callbacks.OnError != nil {
			cmd.OnError(callbacks.OnError)
		}
	}

	// Run async
	return cmd.RunAsync()
}
