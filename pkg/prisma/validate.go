package prisma

import (
	"strings"
)

// ValidateResult holds the result of schema validation
type ValidateResult struct {
	Valid  bool     // True if schema is valid
	Errors []string // List of validation errors
	Output string   // Full output from validate command
}

// Validate runs `npx prisma validate` to check schema validity
func Validate(projectDir string) (*ValidateResult, error) {
	cmd := cmdBuilder.New("npx", "prisma", "validate").WithWorkingDir(projectDir)
	result, err := cmd.RunWithOutput()

	// Parse result
	validateResult := &ValidateResult{
		Output: result.Stdout + result.Stderr,
	}

	// Exit code 0 means validation succeeded
	if err == nil && result.ExitCode == 0 {
		validateResult.Valid = true
		return validateResult, nil
	}

	// Validation failed - parse errors
	validateResult.Valid = false
	validateResult.Errors = parseValidationErrors(result.Stdout, result.Stderr)

	// Return result even if command failed (validation failure is expected behavior)
	return validateResult, nil
}

// parseValidationErrors extracts error messages from prisma validate output
func parseValidationErrors(stdout, stderr string) []string {
	var errors []string
	output := stdout + "\n" + stderr

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Look for error indicators
		if strings.Contains(line, "Error:") ||
			strings.Contains(line, "error:") ||
			strings.Contains(line, "✘") ||
			strings.Contains(line, "×") {
			errors = append(errors, line)
		}
	}

	// If no specific errors found but validation failed, include full output
	if len(errors) == 0 && output != "" {
		// Take first non-empty line as error summary
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "Environment variables") {
				errors = append(errors, line)
				break
			}
		}
	}

	return errors
}
