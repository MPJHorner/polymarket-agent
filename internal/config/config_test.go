package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test default values when no config file exists
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Database.Path != "polytracker.db" {
		t.Errorf("Expected default database path 'polytracker.db', got '%s'", cfg.Database.Path)
	}

	if cfg.UI.Theme != "dracula" {
		t.Errorf("Expected default theme 'dracula', got '%s'", cfg.UI.Theme)
	}
}

func TestEnvironmentOverrides(t *testing.T) {
	os.Setenv("POLYTRACKER_DATABASE_PATH", "test.db")
	defer os.Unsetenv("POLYTRACKER_DATABASE_PATH")

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Database.Path != "test.db" {
		t.Errorf("Expected database path 'test.db' from env, got '%s'", cfg.Database.Path)
	}
}

func TestCreateDefaultConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")
	err = CreateDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to create default config: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created at %s", configPath)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load created config: %v", err)
	}

	if cfg.UI.Theme != "dracula" {
		t.Errorf("Expected theme 'dracula' in created config, got '%s'", cfg.UI.Theme)
	}
}
