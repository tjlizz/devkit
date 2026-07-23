package config

import "testing"

func TestDefault(t *testing.T) {
	t.Setenv("DEVKIT_PORT", "")
	t.Setenv("DEVKIT_DB_PATH", "")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Database.Path != "./devkit.db" {
		t.Errorf("database path = %q, want ./devkit.db", cfg.Database.Path)
	}
}
