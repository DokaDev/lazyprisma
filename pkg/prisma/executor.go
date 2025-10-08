package prisma

import "lazyprisma/pkg/exec"

type Executor struct {
	runner *exec.Runner
}

func NewExecutor() *Executor {
	return &Executor{
		runner: exec.NewRunner(),
	}
}

func (e *Executor) Init() (string, error) {
	return e.runner.RunNPX("prisma", "init")
}

func (e *Executor) Generate() (string, error) {
	return e.runner.RunNPX("prisma", "generate")
}

func (e *Executor) MigrateStatus() (string, error) {
	return e.runner.RunNPX("prisma", "migrate", "status")
}

func (e *Executor) MigrateDiff() (string, error) {
	return e.runner.RunNPX("prisma", "migrate", "diff",
		"--from-schema-datasource", "prisma/schema.prisma",
		"--to-schema-datamodel", "prisma/schema.prisma",
		"--exit-code")
}

func (e *Executor) MigrateDev(name string) (string, error) {
	return e.runner.RunNPX("prisma", "migrate", "dev", "--name", name)
}

func (e *Executor) MigrateDeploy() (string, error) {
	return e.runner.RunNPX("prisma", "migrate", "deploy")
}

func (e *Executor) MigrateReset() (string, error) {
	return e.runner.RunNPX("prisma", "migrate", "reset", "--force")
}

func (e *Executor) MigrateResolve(migrationName, resolveType string) (string, error) {
	return e.runner.RunNPX("prisma", "migrate", "resolve", "--"+resolveType, migrationName)
}

func (e *Executor) Validate() (string, error) {
	return e.runner.RunNPX("prisma", "validate")
}

func (e *Executor) Format() (string, error) {
	return e.runner.RunNPX("prisma", "format")
}

func (e *Executor) Studio() (string, error) {
	return e.runner.RunNPX("prisma", "studio")
}

func (e *Executor) Help() (string, error) {
	return e.runner.RunNPX("prisma", "--help")
}
