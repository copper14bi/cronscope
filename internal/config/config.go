package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Job represents a single cron job to monitor.
type Job struct {
	Name           string        `yaml:"name"`
	Schedule       string        `yaml:"schedule"`
	TimeoutSeconds int           `yaml:"timeout_seconds"`
	GracePeriod    time.Duration `yaml:"-"`
}

// WebhookConfig holds webhook delivery settings.
type WebhookConfig struct {
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Timeout int               `yaml:"timeout_seconds"`
}

// Config is the root configuration structure.
type Config struct {
	Webhook WebhookConfig `yaml:"webhook"`
	Jobs    []Job         `yaml:"jobs"`
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

	for i := range cfg.Jobs {
		if cfg.Jobs[i].TimeoutSeconds > 0 {
			cfg.Jobs[i].GracePeriod = time.Duration(cfg.Jobs[i].TimeoutSeconds) * time.Second
		}
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Webhook.URL == "" {
		return fmt.Errorf("webhook.url is required")
	}
	if len(c.Jobs) == 0 {
		return fmt.Errorf("at least one job must be defined")
	}
	names := make(map[string]struct{}, len(c.Jobs))
	for _, j := range c.Jobs {
		if j.Name == "" {
			return fmt.Errorf("each job must have a name")
		}
		if j.Schedule == "" {
			return fmt.Errorf("job %q must have a schedule", j.Name)
		}
		if _, dup := names[j.Name]; dup {
			return fmt.Errorf("duplicate job name: %q", j.Name)
		}
		names[j.Name] = struct{}{}
	}
	return nil
}
