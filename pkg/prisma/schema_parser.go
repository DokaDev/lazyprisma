package prisma

import (
	"bufio"
	"os"
	"strings"
)

type SchemaInfo struct {
	GeneratorOutput       string
	GeneratorOutputIsSet  bool // Whether output is explicitly set
	DatasourceProvider    string
	DatasourceEnvVar      string // Environment variable name used in datasource url (e.g., DATABASE_URL)
}

type SchemaParser struct{}

func NewSchemaParser() *SchemaParser {
	return &SchemaParser{}
}

func (p *SchemaParser) Parse() (SchemaInfo, error) {
	file, err := os.Open("prisma/schema.prisma")
	if err != nil {
		return SchemaInfo{}, err
	}
	defer file.Close()

	var info SchemaInfo
	scanner := bufio.NewScanner(file)

	inGenerator := false
	inDatasource := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "generator") {
			inGenerator = true
			inDatasource = false
			continue
		}

		if strings.HasPrefix(line, "datasource") {
			inDatasource = true
			inGenerator = false
			continue
		}

		if line == "}" {
			inGenerator = false
			inDatasource = false
			continue
		}

		if inGenerator && strings.Contains(line, "output") {
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				output := strings.TrimSpace(parts[1])
				output = strings.Trim(output, "\"")
				info.GeneratorOutput = output
				info.GeneratorOutputIsSet = true
			}
		}

		if inDatasource && strings.Contains(line, "provider") {
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				provider := strings.TrimSpace(parts[1])
				provider = strings.Trim(provider, "\"")
				info.DatasourceProvider = provider
			}
		}

		if inDatasource && strings.Contains(line, "url") {
			// Extract environment variable name from url = env("DATABASE_URL") format
			if strings.Contains(line, "env(") {
				start := strings.Index(line, "env(\"") + 5
				end := strings.Index(line[start:], "\")")
				if end > 0 {
					info.DatasourceEnvVar = line[start : start+end]
				}
			}
		}
	}

	return info, scanner.Err()
}
