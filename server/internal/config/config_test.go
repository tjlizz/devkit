package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	t.Setenv("DEVKIT_PORT", "")
	t.Setenv("DEVKIT_DB_PATH", "")
	clearAuthEnvironment(t)

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
	if cfg.Auth.JWTSecret == "" {
		t.Error("JWT secret must have a development default")
	}
	if cfg.Auth.JWTExpiryHours != 24 {
		t.Errorf("JWT expiry = %d, want 24", cfg.Auth.JWTExpiryHours)
	}
	if cfg.Auth.SMTP.Port != 587 {
		t.Errorf("SMTP port = %d, want 587", cfg.Auth.SMTP.Port)
	}
}

func TestLoadAuthFromYAMLAndEnvironment(t *testing.T) {
	clearAuthEnvironment(t)
	path := filepath.Join(t.TempDir(), "config.yaml")
	configFile := []byte(`
server:
  port: 8081
database:
  path: test.db
auth:
  jwt_secret: yaml-secret
  jwt_expiry_hours: 12
  activation_base_url: https://yaml.example
  smtp:
    host: smtp.yaml.example
    port: 2525
    username: yaml-user
    password: yaml-password
    from: yaml@example.com
`)
	if err := os.WriteFile(path, configFile, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("DEVKIT_JWT_SECRET", "environment-secret")
	t.Setenv("DEVKIT_SMTP_HOST", "smtp.environment.example")
	t.Setenv("DEVKIT_SMTP_PORT", "465")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Auth.JWTSecret != "environment-secret" {
		t.Errorf("JWT secret = %q, want environment-secret", cfg.Auth.JWTSecret)
	}
	if cfg.Auth.JWTExpiryHours != 12 {
		t.Errorf("JWT expiry = %d, want 12", cfg.Auth.JWTExpiryHours)
	}
	if cfg.Auth.ActivationBaseURL != "https://yaml.example" {
		t.Errorf("activation URL = %q, want YAML value", cfg.Auth.ActivationBaseURL)
	}
	if cfg.Auth.SMTP.Host != "smtp.environment.example" || cfg.Auth.SMTP.Port != 465 {
		t.Errorf("SMTP endpoint = %s:%d, want environment values", cfg.Auth.SMTP.Host, cfg.Auth.SMTP.Port)
	}
	if cfg.Auth.SMTP.Username != "yaml-user" || cfg.Auth.SMTP.Password != "yaml-password" {
		t.Error("SMTP YAML credentials were not loaded")
	}
}

func clearAuthEnvironment(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"DEVKIT_JWT_SECRET",
		"DEVKIT_JWT_EXPIRY_HOURS",
		"DEVKIT_ACTIVATION_BASE_URL",
		"DEVKIT_SMTP_HOST",
		"DEVKIT_SMTP_PORT",
		"DEVKIT_SMTP_USERNAME",
		"DEVKIT_SMTP_PASSWORD",
		"DEVKIT_SMTP_FROM",
	} {
		t.Setenv(name, "")
	}
}
