package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Environment represents a single Vault environment configuration.
type Environment struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
	Token   string `yaml:"token"`
	Prefix  string `yaml:"prefix"`
}

// Config holds the top-level vaultlink configuration.
type Config struct {
	Version      string        `yaml:"version"`
	Environments []Environment `yaml:"environments"`
}

// Load reads and parses a vaultlink config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// Validate checks that the configuration is well-formed.
func (c *Config) Validate() error {
	if len(c.Environments) == 0 {
		return fmt.Errorf("at least one environment must be defined")
	}

	seen := make(map[string]bool)
	for _, env := range c.Environments {
		if env.Name == "" {
			return fmt.Errorf("environment name must not be empty")
		}
		if env.Address == "" {
			return fmt.Errorf("environment %q: address must not be empty", env.Name)
		}
		if seen[env.Name] {
			return fmt.Errorf("duplicate environment name %q", env.Name)
		}
		seen[env.Name] = true
	}
	return nil
}

// FindEnvironment returns the environment with the given name, or an error.
func (c *Config) FindEnvironment(name string) (*Environment, error) {
	for i := range c.Environments {
		if c.Environments[i].Name == name {
			return &c.Environments[i], nil
		}
	}
	return nil, fmt.Errorf("environment %q not found in config", name)
}
