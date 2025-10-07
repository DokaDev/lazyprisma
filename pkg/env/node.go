package env

import (
	"strings"

	"github.com/DokaDev/lazyprisma/pkg/exec"
)

type NodeChecker struct {
	runner *exec.Runner
}

func NewNodeChecker() *NodeChecker {
	return &NodeChecker{
		runner: exec.NewRunner(),
	}
}

func (n *NodeChecker) CheckNode() string {
	output, err := n.runner.Run("node", "--version")
	if err != nil {
		return "Not installed"
	}
	return strings.TrimSpace(output)
}

func (n *NodeChecker) CheckNPM() string {
	output, err := n.runner.Run("npm", "--version")
	if err != nil {
		return "Not installed"
	}
	return strings.TrimSpace(output)
}
