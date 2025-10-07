package exec

import (
	"os/exec"
)

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (r *Runner) RunNPX(args ...string) (string, error) {
	return r.Run("npx", args...)
}
