package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level driftwatch daemon configuration.
type Config struct {
	PollInterval time.Duration `yaml:"poll_interval"`
	LogLevel     string        `yaml:"log_level"`
	Services     []Service     `yaml:"services"`
}

// Service describes a single service to watch for configuration drift.
type Service struct {
	Name       string `yaml:"name"`
	DeclaredAt string `yaml:"declared_at"` // path or URL to declared state (e.g. git path)
	Endpoint   string `yaml:"endpoint"`    // running service config endpoint
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("config: decode %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: validation failed: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.PollInterval <= 0 {
		c.PollInterval = 30 * time.Second
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	for i, svc := range c.Services {
		if svc.Name == "" {
			return fmt.Errorf("service[%d]: name is required", i)
		}
		if svc.DeclaredAt == "" {
			return fmt.Errorf("service %q: declared_at is required", svc.Name)
		}
		if svc.Endpoint == "" {
			return fmt.Errorf("service %q: endpoint is required", svc.Name)
		}
	}
	return nil
}
