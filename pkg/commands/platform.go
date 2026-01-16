package commands

import "runtime"

// Platform holds platform-specific configurations
type Platform struct {
	OS       string
	Shell    string
	ShellArg string
}

// NewPlatform creates a platform configuration for the current OS
func NewPlatform() *Platform {
	os := runtime.GOOS

	shell := "sh"
	shellArg := "-c"

	if os == "windows" {
		shell = "cmd"
		shellArg = "/c"
	}

	return &Platform{
		OS:       os,
		Shell:    shell,
		ShellArg: shellArg,
	}
}

// GetShell returns the shell and its command argument
func (p *Platform) GetShell() (string, string) {
	return p.Shell, p.ShellArg
}
