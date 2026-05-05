package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Job defines a single monitored cron job.
type Job struct {
	Name     string        `yaml:"name"`
	Schedule string        `yaml:"schedule"`
	Grace    time.Duration `yaml:"grace"`
	Alert    AlertConfig   `yaml:"alert"`
}

// AlertConfig holds per-job alert settings.
type AlertConfig struct {
	Email   string `yaml:"email"`
	Webhook string `yaml:"webhook"`
}

// Config is the top-level configuration structure.
type Config struct {
	ListenAddr string      `yaml:"listen_addr"`
	LogLevel   string      `yaml:"log_level"`
	Jobs       []Job       `yaml:"jobs"`
	Alert      AlertConfig `yaml:"alert"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.Jobs) == 0 {
		return fmt.Errorf("no jobs defined")
	}
	seen := make(map[string]bool)
	for _, j := range c.Jobs {
		if j.Name == "" {
			return fmt.Errorf("job missing name")
		}
		if j.Schedule == "" {
			return fmt.Errorf("job %q missing schedule", j.Name)
		}
		if seen[j.Name] {
			return fmt.Errorf("duplicate job name %q", j.Name)
		}
		seen[j.Name] = true
	}
	return nil
}
