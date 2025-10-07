package env

import (
	"bufio"
	"os"
	"strings"
)

type DotEnvReader struct {
	envVarName string // Environment variable name defined in schema.prisma
}

func NewDotEnvReader(envVarName string) *DotEnvReader {
	// Use default value if environment variable name is not provided
	if envVarName == "" {
		envVarName = "DATABASE_URL"
	}
	return &DotEnvReader{
		envVarName: envVarName,
	}
}

func (d *DotEnvReader) GetDatabaseURL() string {
	return d.GetEnvVar(d.envVarName)
}

func (d *DotEnvReader) GetEnvVar(key string) string {
	// Check environment variables first
	if url := os.Getenv(key); url != "" {
		return url
	}

	// Read from .env file
	file, err := os.Open(".env")
	if err != nil {
		return "Not configured"
	}
	defer file.Close()

	prefix := key + "="
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, prefix) {
			value := strings.TrimPrefix(line, prefix)
			value = strings.Trim(value, "\"'")
			return value
		}
	}

	return "Not configured"
}

func (d *DotEnvReader) MaskDatabaseURL(url string) string {
	if url == "Not configured" {
		return url
	}

	parts := strings.Split(url, "@")
	if len(parts) < 2 {
		return url
	}

	userPart := parts[0]
	hostPart := parts[1]

	if strings.Contains(userPart, ":") {
		credParts := strings.SplitN(userPart, ":", 2)
		if len(credParts) == 2 {
			userPart = credParts[0] + ":***"
		}
	}

	return userPart + "@" + hostPart
}
