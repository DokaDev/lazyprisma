package prisma

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Datasource represents a Prisma datasource configuration
type Datasource struct {
	Provider    string // Database provider (postgresql, mysql, sqlite, etc.)
	URL         string // Database connection URL
	EnvVarName  string // Environment variable name (e.g., "DATABASE_URL")
	IsHardcoded bool   // True if URL is hardcoded in schema/config
}

// GetProvider extracts only the provider from schema files
// This is useful when URL resolution fails but we still want to show the provider
func GetProvider(projectDir string) (string, error) {
	// Both v7+ and v7- projects have provider in schema.prisma
	return extractProviderFromSchema(projectDir)
}

// GetEnvVarName extracts the environment variable name used for database URL
// Returns the env var name even if the variable is not set
func GetEnvVarName(projectDir string) (string, error) {
	// Check if this is a v7+ project
	configPath := filepath.Join(projectDir, ConfigFileName)
	if _, err := os.Stat(configPath); err == nil {
		// v7+: Check prisma.config.ts
		return extractEnvVarFromConfig(configPath)
	}

	// v7-: Check schema.prisma
	return extractEnvVarFromSchema(projectDir)
}

// GetDatasource extracts datasource information from schema.prisma or prisma.config.ts
func GetDatasource(projectDir string) (*Datasource, error) {
	// Check if this is a v7+ project (has prisma.config.ts)
	configPath := filepath.Join(projectDir, ConfigFileName)
	isV7Plus := false
	if _, err := os.Stat(configPath); err == nil {
		isV7Plus = true
	}

	ds := &Datasource{}

	if isV7Plus {
		// v7+: Get URL from prisma.config.ts, provider from schema.prisma
		url, envVar, isHardcoded, err := extractURLFromConfig(projectDir, configPath)
		if err != nil {
			return nil, err
		}
		ds.URL = url
		ds.EnvVarName = envVar
		ds.IsHardcoded = isHardcoded

		provider, err := extractProviderFromSchema(projectDir)
		if err != nil {
			return nil, err
		}
		ds.Provider = provider
	} else {
		// v7-: Get both from schema.prisma
		provider, url, envVar, isHardcoded, err := extractDatasourceFromSchema(projectDir)
		if err != nil {
			return nil, err
		}
		ds.Provider = provider
		ds.URL = url
		ds.EnvVarName = envVar
		ds.IsHardcoded = isHardcoded
	}

	if ds.Provider == "" || ds.URL == "" {
		return nil, fmt.Errorf("incomplete datasource configuration")
	}

	return ds, nil
}

// extractEnvVarFromConfig extracts only the env var name from prisma.config.ts
func extractEnvVarFromConfig(configPath string) (string, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to open config: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	envRegex := regexp.MustCompile(`url:\s*env\(['"]([^'"]+)['"]\)`)

	for scanner.Scan() {
		line := scanner.Text()
		if match := envRegex.FindStringSubmatch(line); match != nil {
			return match[1], nil
		}
	}

	return "", fmt.Errorf("env var not found in config")
}

// extractURLFromConfig extracts database URL from prisma.config.ts
// Returns: (url, envVarName, isHardcoded, error)
func extractURLFromConfig(projectDir, configPath string) (string, string, bool, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return "", "", false, fmt.Errorf("failed to open config: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	envRegex := regexp.MustCompile(`url:\s*env\(['"]([^'"]+)['"]\)`)
	hardcodedRegex := regexp.MustCompile(`url:\s*['"]([^'"]+)['"]`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for env() usage
		if match := envRegex.FindStringSubmatch(line); match != nil {
			envVar := match[1]
			url := resolveEnvVar(projectDir, envVar)
			return url, envVar, false, nil
		}

		// Check for hardcoded URL
		if match := hardcodedRegex.FindStringSubmatch(line); match != nil {
			url := match[1]
			return url, "", true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", false, fmt.Errorf("failed to read config: %w", err)
	}

	return "", "", false, fmt.Errorf("datasource.url not found in config")
}

// extractEnvVarFromSchema extracts only the env var name from schema.prisma
func extractEnvVarFromSchema(projectDir string) (string, error) {
	schemaPath := filepath.Join(projectDir, SchemaDirName, SchemaFileName)
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		return "", fmt.Errorf("schema.prisma not found")
	}

	file, err := os.Open(schemaPath)
	if err != nil {
		return "", fmt.Errorf("failed to open schema: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inDatasource := false
	envRegex := regexp.MustCompile(`env\("([^"]+)"\)`)
	urlRegex := regexp.MustCompile(`url\s*=\s*(.+)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "datasource") {
			inDatasource = true
			continue
		}

		if inDatasource && line == "}" {
			break
		}

		if inDatasource {
			// Check for url field
			if match := urlRegex.FindStringSubmatch(line); match != nil {
				urlValue := strings.TrimSpace(match[1])
				// Extract env var name
				if envMatch := envRegex.FindStringSubmatch(urlValue); envMatch != nil {
					return envMatch[1], nil
				}
			}
		}
	}

	return "", fmt.Errorf("env var not found in schema")
}

// extractProviderFromSchema extracts provider from schema.prisma
func extractProviderFromSchema(projectDir string) (string, error) {
	schemaPath := filepath.Join(projectDir, SchemaDirName, SchemaFileName)
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		return "", fmt.Errorf("schema.prisma not found")
	}

	file, err := os.Open(schemaPath)
	if err != nil {
		return "", fmt.Errorf("failed to open schema: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inDatasource := false
	providerRegex := regexp.MustCompile(`provider\s*=\s*"([^"]+)"`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "datasource") {
			inDatasource = true
			continue
		}

		if inDatasource && line == "}" {
			break
		}

		if inDatasource {
			if match := providerRegex.FindStringSubmatch(line); match != nil {
				return match[1], nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read schema: %w", err)
	}

	return "", fmt.Errorf("provider not found in schema")
}

// extractDatasourceFromSchema extracts both provider and URL from schema.prisma (v7-)
// Returns: (provider, url, envVarName, isHardcoded, error)
func extractDatasourceFromSchema(projectDir string) (string, string, string, bool, error) {
	schemaPath := filepath.Join(projectDir, SchemaDirName, SchemaFileName)
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		return "", "", "", false, fmt.Errorf("schema.prisma not found")
	}

	file, err := os.Open(schemaPath)
	if err != nil {
		return "", "", "", false, fmt.Errorf("failed to open schema: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inDatasource := false
	provider := ""
	url := ""
	envVarName := ""
	isHardcoded := false

	providerRegex := regexp.MustCompile(`provider\s*=\s*"([^"]+)"`)
	urlRegex := regexp.MustCompile(`url\s*=\s*(.+)`)
	envRegex := regexp.MustCompile(`env\("([^"]+)"\)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "datasource") {
			inDatasource = true
			continue
		}

		if inDatasource && line == "}" {
			break
		}

		if inDatasource {
			// Extract provider
			if match := providerRegex.FindStringSubmatch(line); match != nil {
				provider = match[1]
			}

			// Extract url
			if match := urlRegex.FindStringSubmatch(line); match != nil {
				urlValue := strings.TrimSpace(match[1])

				// Check if it's an env variable reference
				if envMatch := envRegex.FindStringSubmatch(urlValue); envMatch != nil {
					envVarName = envMatch[1]
					url = resolveEnvVar(projectDir, envVarName)
					isHardcoded = false
				} else {
					// Direct URL (hardcoded, remove quotes)
					url = strings.Trim(urlValue, `"`)
					isHardcoded = true
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", "", false, fmt.Errorf("failed to read schema: %w", err)
	}

	return provider, url, envVarName, isHardcoded, nil
}

// resolveEnvVar resolves an environment variable following Prisma's resolution order
func resolveEnvVar(projectDir, envVar string) string {
	// 1. Check OS environment variables first
	if val := os.Getenv(envVar); val != "" {
		return val
	}

	// 2. Check .env in project root
	if val := readEnvFile(filepath.Join(projectDir, ".env"), envVar); val != "" {
		return val
	}

	// 3. Check .env in schema directory
	if val := readEnvFile(filepath.Join(projectDir, SchemaDirName, ".env"), envVar); val != "" {
		return val
	}

	// 4. Check .env in prisma directory
	if val := readEnvFile(filepath.Join(projectDir, "prisma", ".env"), envVar); val != "" {
		return val
	}

	return ""
}

// readEnvFile reads a specific environment variable from a .env file
func readEnvFile(path, envVar string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	envRegex := regexp.MustCompile(`^\s*` + regexp.QuoteMeta(envVar) + `\s*=\s*(.+)`)

	for scanner.Scan() {
		line := scanner.Text()
		if match := envRegex.FindStringSubmatch(line); match != nil {
			// Remove quotes if present
			value := strings.TrimSpace(match[1])
			value = strings.Trim(value, `"'`)
			return value
		}
	}

	return ""
}

// MaskPassword masks the password in a database URL with asterisks
func MaskPassword(dbURL string) string {
	if dbURL == "" {
		return ""
	}

	// Use regex to find and replace password in the URL string
	// Pattern: ://username:password@ -> ://username:****@
	passwordRegex := regexp.MustCompile(`(://[^:]+:)([^@]+)(@)`)

	// Check if password exists in URL
	if !passwordRegex.MatchString(dbURL) {
		return dbURL // No password to mask
	}

	// Replace password with asterisks
	maskedURL := passwordRegex.ReplaceAllString(dbURL, "${1}****${3}")
	return maskedURL
}
