package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	AppName       = "lazyprisma"
	ConfigDirName = "lazyprisma"
	ConfigFile    = "config.yaml"
)

// Config holds application configuration
type Config struct {
	Scan ScanConfig `yaml:"scan"`
}

// ScanConfig holds project scanning settings
type ScanConfig struct {
	MaxDepth    int      `yaml:"maxDepth"`
	ExcludeDirs []string `yaml:"excludeDirs"`
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		Scan: ScanConfig{
			MaxDepth:    10,
			ExcludeDirs: []string{}, // Additional excludes (defaults are in prisma.DefaultExcludeDirs)
		},
	}
}

// ConfigDir returns the config directory path
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", ConfigDirName), nil
}

// ConfigPath returns the full path to the config file
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFile), nil
}

// Load loads configuration from ~/.config/lazyprisma/config.yaml
// Returns default config if file doesn't exist
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return Default(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}

	cfg := Default() // Start with defaults
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save saves configuration to ~/.config/lazyprisma/config.yaml
func Save(cfg *Config) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// EnsureConfigFile creates the config file with defaults if it doesn't exist
func EnsureConfigFile() error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path, err := ConfigPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Write a nicely formatted default config
		defaultConfig := `# LazyPrisma Configuration
scan:
  # Maximum directory depth for scanning (0 = unlimited)
  maxDepth: 10
  # Additional directories to exclude (defaults like node_modules are always excluded)
  excludeDirs:
    # - /full/path/to/exclude
    # - dirname-to-exclude
`
		return os.WriteFile(path, []byte(defaultConfig), 0644)
	}

	return nil
}
