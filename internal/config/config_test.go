package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "croncheck-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	path := writeTempConfig(t, `
listen_addr: ":8080"
log_level: info
jobs:
  - name: backup
    schedule: "0 2 * * *"
    grace: 10m
    alert:
      email: ops@example.com
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(cfg.Jobs))
	}
	if cfg.Jobs[0].Grace != 10*time.Minute {
		t.Errorf("expected grace 10m, got %v", cfg.Jobs[0].Grace)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_NoJobs(t *testing.T) {
	path := writeTempConfig(t, `listen_addr: ":8080"
jobs: []
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for empty jobs")
	}
}

func TestLoad_DuplicateJobName(t *testing.T) {
	path := writeTempConfig(t, `
jobs:
  - name: backup
    schedule: "0 2 * * *"
  - name: backup
    schedule: "0 3 * * *"
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate job name")
	}
}
