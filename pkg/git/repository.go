package git

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

// GitInfo holds Git repository information
type GitInfo struct {
	RepositoryName string
	BranchName     string
	IsRepository   bool
}

// GetGitInfo returns Git repository information for the given directory
func GetGitInfo(dir string) *GitInfo {
	info := &GitInfo{}

	// Find git repository root (walks up parent directories)
	gitRoot := findGitRoot(dir)
	if gitRoot == "" {
		info.IsRepository = false
		return info
	}

	info.IsRepository = true

	// Get repository name
	info.RepositoryName = getRepositoryName(gitRoot)

	// Get current branch
	info.BranchName = getCurrentBranch(gitRoot)

	return info
}

// findGitRoot finds the git repository root by walking up parent directories
func findGitRoot(dir string) string {
	currentDir := dir

	// Walk up to root directory
	for {
		gitDir := filepath.Join(currentDir, ".git")
		if stat, err := os.Stat(gitDir); err == nil && stat.IsDir() {
			return currentDir
		}

		// Move to parent directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached filesystem root
			break
		}
		currentDir = parentDir
	}

	return ""
}

// getRepositoryName returns the repository name
func getRepositoryName(dir string) string {
	// Try to get from remote URL
	cmd := cmdBuilder.New("git", "remote", "get-url", "origin").WithWorkingDir(dir)
	result, err := cmd.RunWithOutput()
	if err == nil && result.Stdout != "" {
		// Parse repository name from URL
		url := strings.TrimSpace(result.Stdout)
		return parseRepoNameFromURL(url)
	}

	// Fallback: use directory name
	return filepath.Base(dir)
}

// getCurrentBranch returns the current branch name
func getCurrentBranch(dir string) string {
	cmd := cmdBuilder.New("git", "branch", "--show-current").WithWorkingDir(dir)
	result, err := cmd.RunWithOutput()
	if err == nil && result.Stdout != "" {
		return strings.TrimSpace(result.Stdout)
	}

	// Fallback: try reading .git/HEAD
	headFile := filepath.Join(dir, ".git", "HEAD")
	content, err := os.ReadFile(headFile)
	if err == nil {
		head := strings.TrimSpace(string(content))
		if strings.HasPrefix(head, "ref: refs/heads/") {
			return strings.TrimPrefix(head, "ref: refs/heads/")
		}
	}

	return "unknown"
}

// parseRepoNameFromURL extracts repository name from git URL
func parseRepoNameFromURL(url string) string {
	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Extract last part of URL
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return url
}

// IsFileModified checks if a file has been modified in the working tree
// Returns true if the file has any changes (staged or unstaged)
func IsFileModified(dir, filePath string) bool {
	// Find git repository root
	gitRoot := findGitRoot(dir)
	if gitRoot == "" {
		return false
	}

	// Get relative path from git root
	relPath, err := filepath.Rel(gitRoot, filePath)
	if err != nil {
		return false
	}

	// Run git status --porcelain to check file status
	cmd := cmdBuilder.New("git", "status", "--porcelain", relPath).WithWorkingDir(gitRoot)
	result, err := cmd.RunWithOutput()
	if err != nil {
		return false
	}

	// If output is not empty, file has changes
	return strings.TrimSpace(result.Stdout) != ""
}
