package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Job represents a single monitored cron job entry.
type Job struct {
	Name     string `yaml:"name"`
	Schedule string `yaml:"schedule"`
	Alert    string `yaml:"alert,omitempty"`
}

// Config is the top-level configuration structure.
type Config struct {
	Jobs           []Job  `yaml:"jobs"`
	AlertEmail     string `yaml:"alert_email,omitempty"`
	SMTPHost       string `yaml:"smtp_host,omitempty"`
	SMTPPort       int    `yaml:"smtp_port,omitempty"`
	CheckIntervalS int    `yaml:"check_interval_seconds,omitempty"`
}

// Load reads and validates a YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	if cfg.CheckIntervalS == 0 {
		cfg.CheckIntervalS = 60
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if len(cfg.Jobs) == 0 {
		return errors.New("config must define at least one job")
	}

	seen := make(map[string]struct{}, len(cfg.Jobs))
	for i, job := range cfg.Jobs {
		if job.Name == "" {
			return fmt.Errorf("job[%d]: name is required", i)
		}
		if job.Schedule == "" {
			return fmt.Errorf("job %q: schedule is required", job.Name)
		}
		if _, dup := seen[job.Name]; dup {
			return fmt.Errorf("duplicate job name: %q", job.Name)
		}
		seen[job.Name] = struct{}{}
	}
	return nil
}
