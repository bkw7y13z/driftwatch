package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "driftwatch-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTempConfig(t, `
poll_interval: 10s
log_level: debug
services:
  - name: auth-service
    declared_at: ./configs/auth.yaml
    endpoint: http://localhost:8081/config
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != 10*time.Second {
		t.Errorf("poll_interval: got %v, want 10s", cfg.PollInterval)
	}
	if len(cfg.Services) != 1 {
		t.Fatalf("services: got %d, want 1", len(cfg.Services))
	}
	if cfg.Services[0].Name != "auth-service" {
		t.Errorf("service name: got %q, want auth-service", cfg.Services[0].Name)
	}
}

func TestLoad_DefaultsApplied(t *testing.T) {
	path := writeTempConfig(t, `
services:
  - name: svc
    declared_at: ./cfg.yaml
    endpoint: http://localhost:9000/config
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != 30*time.Second {
		t.Errorf("default poll_interval: got %v, want 30s", cfg.PollInterval)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("default log_level: got %q, want info", cfg.LogLevel)
	}
}

func TestLoad_MissingServiceName(t *testing.T) {
	path := writeTempConfig(t, `
services:
  - declared_at: ./cfg.yaml
    endpoint: http://localhost:9000/config
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing service name, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
