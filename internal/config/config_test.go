package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cronscope-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTempConfig(t, `
webhook:
  url: "https://hooks.example.com/alert"
  timeout_seconds: 5
jobs:
  - name: backup
    schedule: "0 2 * * *"
    timeout_seconds: 3600
  - name: cleanup
    schedule: "*/15 * * * *"
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Webhook.URL != "https://hooks.example.com/alert" {
		t.Errorf("unexpected webhook URL: %s", cfg.Webhook.URL)
	}
	if len(cfg.Jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(cfg.Jobs))
	}
	if cfg.Jobs[0].GracePeriod != 3600*time.Second {
		t.Errorf("expected grace period 3600s, got %v", cfg.Jobs[0].GracePeriod)
	}
}

func TestLoad_MissingWebhookURL(t *testing.T) {
	path := writeTempConfig(t, `
webhook:
  url: ""
jobs:
  - name: backup
    schedule: "0 2 * * *"
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing webhook URL")
	}
}

func TestLoad_DuplicateJobName(t *testing.T) {
	path := writeTempConfig(t, `
webhook:
  url: "https://hooks.example.com/alert"
jobs:
  - name: backup
    schedule: "0 2 * * *"
  - name: backup
    schedule: "*/5 * * * *"
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate job name")
	}
}

func TestLoad_NoJobs(t *testing.T) {
	path := writeTempConfig(t, `
webhook:
  url: "https://hooks.example.com/alert"
jobs: []
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error when no jobs defined")
	}
}
