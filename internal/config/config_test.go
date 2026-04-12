package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/user/portwatch/internal/config"
)

func TestDefaults(t *testing.T) {
	cfg := config.Defaults()

	if cfg.Interval != 5*time.Second {
		t.Errorf("expected default interval 5s, got %s", cfg.Interval)
	}
	if cfg.ProcNet.TCP != "/proc/net/tcp" {
		t.Errorf("unexpected default TCP path: %s", cfg.ProcNet.TCP)
	}
	if cfg.AlertHook.Timeout != 5*time.Second {
		t.Errorf("expected default hook timeout 5s, got %s", cfg.AlertHook.Timeout)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/portwatch.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_ValidFile(t *testing.T) {
	content := `
interval: 10s
log_file: /var/log/portwatch.log
alert_hook:
  url: http://example.com/hook
  timeout: 3s
`
	tmp, err := os.CreateTemp("", "portwatch-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmp.Close()

	cfg, err := config.Load(tmp.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != 10*time.Second {
		t.Errorf("expected 10s interval, got %s", cfg.Interval)
	}
	if cfg.LogFile != "/var/log/portwatch.log" {
		t.Errorf("unexpected log_file: %s", cfg.LogFile)
	}
	if cfg.AlertHook.URL != "http://example.com/hook" {
		t.Errorf("unexpected hook URL: %s", cfg.AlertHook.URL)
	}
	if cfg.AlertHook.Timeout != 3*time.Second {
		t.Errorf("expected hook timeout 3s, got %s", cfg.AlertHook.Timeout)
	}
	// ProcNet should retain defaults when not specified
	if cfg.ProcNet.TCP != "/proc/net/tcp" {
		t.Errorf("default TCP path overwritten: %s", cfg.ProcNet.TCP)
	}
}

func TestLoad_UnknownField(t *testing.T) {
	content := "unknown_key: value\n"
	tmp, err := os.CreateTemp("", "portwatch-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	tmp.WriteString(content)
	tmp.Close()

	_, err = config.Load(tmp.Name())
	if err == nil {
		t.Fatal("expected error for unknown field, got nil")
	}
}
