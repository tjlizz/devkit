package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

const (
	defaultPort          = 8080
	defaultDatabasePath  = "./devkit.db"
	defaultJWTSecret     = "devkit-development-secret"
	defaultJWTExpiry     = 24
	defaultSMTPPort      = 587
	defaultActivationURL = "http://localhost:8080"
	defaultAvatarDir     = "./uploads/avatars"
	defaultArtifactDir   = "./uploads/artifacts"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type AuthConfig struct {
	JWTSecret         string     `yaml:"jwt_secret"`
	JWTExpiryHours    int        `yaml:"jwt_expiry_hours"`
	ActivationBaseURL string     `yaml:"activation_base_url"`
	AvatarUploadDir   string     `yaml:"avatar_upload_dir"`
	ArtifactUploadDir string     `yaml:"artifact_upload_dir"`
	SMTP              SMTPConfig `yaml:"smtp"`
}

type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

func Default() Config {
	return Config{
		Server: ServerConfig{
			Port: defaultPort,
		},
		Database: DatabaseConfig{
			Path: defaultDatabasePath,
		},
		Auth: AuthConfig{
			JWTSecret:         defaultJWTSecret,
			JWTExpiryHours:    defaultJWTExpiry,
			ActivationBaseURL: defaultActivationURL,
			AvatarUploadDir:   defaultAvatarDir,
			ArtifactUploadDir: defaultArtifactDir,
			SMTP: SMTPConfig{
				Port: defaultSMTPPort,
			},
		},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return Config{}, fmt.Errorf("read config file: %w", err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse config file: %w", err)
		}
	}

	if port := os.Getenv("DEVKIT_PORT"); port != "" {
		value, err := strconv.Atoi(port)
		if err != nil {
			return Config{}, fmt.Errorf("parse DEVKIT_PORT: %w", err)
		}
		cfg.Server.Port = value
	}
	if path := os.Getenv("DEVKIT_DB_PATH"); path != "" {
		cfg.Database.Path = path
	}
	if secret := os.Getenv("DEVKIT_JWT_SECRET"); secret != "" {
		cfg.Auth.JWTSecret = secret
	}
	if expiry := os.Getenv("DEVKIT_JWT_EXPIRY_HOURS"); expiry != "" {
		value, err := strconv.Atoi(expiry)
		if err != nil {
			return Config{}, fmt.Errorf("parse DEVKIT_JWT_EXPIRY_HOURS: %w", err)
		}
		cfg.Auth.JWTExpiryHours = value
	}
	if baseURL := os.Getenv("DEVKIT_ACTIVATION_BASE_URL"); baseURL != "" {
		cfg.Auth.ActivationBaseURL = baseURL
	}
	if uploadDir := os.Getenv("DEVKIT_AVATAR_UPLOAD_DIR"); uploadDir != "" {
		cfg.Auth.AvatarUploadDir = uploadDir
	}
	if uploadDir := os.Getenv("DEVKIT_ARTIFACT_UPLOAD_DIR"); uploadDir != "" {
		cfg.Auth.ArtifactUploadDir = uploadDir
	}
	if host := os.Getenv("DEVKIT_SMTP_HOST"); host != "" {
		cfg.Auth.SMTP.Host = host
	}
	if port := os.Getenv("DEVKIT_SMTP_PORT"); port != "" {
		value, err := strconv.Atoi(port)
		if err != nil {
			return Config{}, fmt.Errorf("parse DEVKIT_SMTP_PORT: %w", err)
		}
		cfg.Auth.SMTP.Port = value
	}
	if username := os.Getenv("DEVKIT_SMTP_USERNAME"); username != "" {
		cfg.Auth.SMTP.Username = username
	}
	if password := os.Getenv("DEVKIT_SMTP_PASSWORD"); password != "" {
		cfg.Auth.SMTP.Password = password
	}
	if from := os.Getenv("DEVKIT_SMTP_FROM"); from != "" {
		cfg.Auth.SMTP.From = from
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535: %d", c.Server.Port)
	}
	if c.Database.Path == "" {
		return errors.New("database path must not be empty")
	}
	if c.Auth.JWTSecret == "" {
		return errors.New("auth JWT secret must not be empty")
	}
	if c.Auth.JWTExpiryHours < 1 {
		return errors.New("auth JWT expiry hours must be positive")
	}
	if c.Auth.ActivationBaseURL == "" {
		return errors.New("auth activation base URL must not be empty")
	}
	if c.Auth.AvatarUploadDir == "" {
		return errors.New("auth avatar upload dir must not be empty")
	}
	if c.Auth.ArtifactUploadDir == "" {
		return errors.New("auth artifact upload dir must not be empty")
	}
	if c.Auth.SMTP.Port < 1 || c.Auth.SMTP.Port > 65535 {
		return fmt.Errorf("SMTP port must be between 1 and 65535: %d", c.Auth.SMTP.Port)
	}
	return nil
}
