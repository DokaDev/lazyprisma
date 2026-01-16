package node

import (
	"strings"

	"github.com/dokadev/lazyprisma/pkg/commands"
)

var cmdBuilder *commands.CommandBuilder

func init() {
	platform := commands.NewPlatform()
	cmdBuilder = commands.NewCommandBuilder(platform)
}

// VersionInfo holds Node.js version information
type VersionInfo struct {
	Version string
}

// IsInstalled checks if Node.js is installed
func IsInstalled() bool {
	cmd := cmdBuilder.New("node", "--version")
	result, err := cmd.RunWithOutput()
	return err == nil && result.ExitCode == 0
}

// GetVersion returns the installed Node.js version
func GetVersion() (*VersionInfo, error) {
	cmd := cmdBuilder.New("node", "--version")
	result, err := cmd.RunWithOutput()
	if err != nil {
		return nil, err
	}

	version := strings.TrimSpace(result.Stdout)
	version = strings.TrimPrefix(version, "v")

	return &VersionInfo{
		Version: version,
	}, nil
}
