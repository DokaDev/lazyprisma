package prisma

import (
	"os"
	"strings"

	"lazyprisma/pkg/exec"
)

type CLIChecker struct {
	runner *exec.Runner
}

func NewCLIChecker() *CLIChecker {
	return &CLIChecker{
		runner: exec.NewRunner(),
	}
}

func (c *CLIChecker) Check() (available bool, version string) {
	output, err := c.runner.RunNPX("prisma", "--version")
	if err != nil {
		return false, "Not installed"
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "prisma") && strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				version = strings.TrimSpace(parts[1])
				return true, version
			}
		}
	}

	return true, "Unknown"
}

func (c *CLIChecker) IsLocalInstalled() bool {
	// Check if Prisma is installed locally in node_modules
	paths := []string{
		"node_modules/.bin/prisma",
		"node_modules/prisma/package.json",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}
