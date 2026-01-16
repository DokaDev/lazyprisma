package prisma

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dokadev/lazyprisma/pkg/commands"
)

var cmdBuilder *commands.CommandBuilder

func init() {
	platform := commands.NewPlatform()
	cmdBuilder = commands.NewCommandBuilder(platform)
}

// VersionInfo holds Prisma version information
type VersionInfo struct {
	Version  string
	IsGlobal bool
}

// GetVersion returns the Prisma version
// It checks local installation first, then falls back to global
func GetVersion(projectDir string) (*VersionInfo, error) {
	// Check if prisma is installed locally (check up to 3 parent directories for monorepo support)
	isLocal := isPrismaInstalledLocally(projectDir)

	// Use npx to get version (it automatically prefers local over global)
	cmd := cmdBuilder.New("npx", "prisma", "--version").WithWorkingDir(projectDir)
	result, err := cmd.RunWithOutput()
	if err != nil {
		// Fallback to global prisma command
		version, err := getPrismaVersionFromPath("prisma")
		if err != nil {
			return nil, err
		}
		return &VersionInfo{
			Version:  version,
			IsGlobal: true,
		}, nil
	}

	// Parse version from output
	version := parseVersionFromOutput(result.Stdout)
	if version == "" {
		return nil, nil
	}

	return &VersionInfo{
		Version:  version,
		IsGlobal: !isLocal,
	}, nil
}

// isPrismaInstalledLocally checks if Prisma is installed locally
// Searches current directory and up to 3 parent directories (for monorepo support)
func isPrismaInstalledLocally(projectDir string) bool {
	currentDir := projectDir

	// Check up to 3 parent directories
	for i := 0; i < 4; i++ {
		prismaDir := filepath.Join(currentDir, "node_modules", "prisma")
		if stat, err := os.Stat(prismaDir); err == nil && stat.IsDir() {
			return true
		}

		// Move to parent directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached root
			break
		}
		currentDir = parentDir
	}

	return false
}

// parseVersionFromOutput parses Prisma version from command output
func parseVersionFromOutput(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "prisma") && strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// getPrismaVersionFromPath gets version from a specific prisma binary path
func getPrismaVersionFromPath(prismaPath string) (string, error) {
	cmd := cmdBuilder.New(prismaPath, "--version")
	result, err := cmd.RunWithOutput()
	if err != nil {
		return "", err
	}

	// Parse version from output
	// Example output: "prisma : 5.22.0"
	lines := strings.Split(result.Stdout, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "prisma") && strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				version := strings.TrimSpace(parts[1])
				return version, nil
			}
		}
	}

	return "", nil
}
