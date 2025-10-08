package prisma

import (
	"lazyprisma/pkg/env"
)

type Status struct {
	CLIAvailable bool
	Version      string
	IsGlobal     bool
	SchemaExists bool
	SchemaInfo   SchemaInfo
	DatabaseURL  string
	NodeVersion  string
	NPMVersion   string
	Migrations   []Migration
}

func GetStatus() Status {
	cliChecker := NewCLIChecker()
	schemaChecker := NewSchemaChecker()
	nodeChecker := env.NewNodeChecker()

	available, version := cliChecker.Check()
	schemaExists := schemaChecker.Check()

	// Check if Prisma is globally installed
	isGlobal := false
	if available && !cliChecker.IsLocalInstalled() {
		isGlobal = true
	}

	nodeVersion := nodeChecker.CheckNode()
	npmVersion := nodeChecker.CheckNPM()

	var schemaInfo SchemaInfo
	var migrations []Migration
	var databaseURL string

	if schemaExists {
		parser := NewSchemaParser()
		parsed, err := parser.Parse()
		if err == nil {
			schemaInfo = parsed
		}

		migrationReader := NewMigrationReader()
		migrations = migrationReader.GetMigrations()

		// Use environment variable name defined in schema (default: DATABASE_URL)
		envVarName := schemaInfo.DatasourceEnvVar
		envReader := env.NewDotEnvReader(envVarName)
		databaseURL = envReader.GetDatabaseURL()
	} else {
		// Use default value if schema does not exist
		envReader := env.NewDotEnvReader("")
		databaseURL = envReader.GetDatabaseURL()
	}

	return Status{
		CLIAvailable: available,
		Version:      version,
		IsGlobal:     isGlobal,
		SchemaExists: schemaExists,
		SchemaInfo:   schemaInfo,
		NodeVersion:  nodeVersion,
		NPMVersion:   npmVersion,
		DatabaseURL:  databaseURL,
		Migrations:   migrations,
	}
}
