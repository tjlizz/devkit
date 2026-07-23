package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

const (
	defaultPort         = 8080
	defaultDatabasePath = "./devkit.db"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

func Default() Config {
	return Config{
		Server: ServerConfig{
			Port: defaultPort,
		},
		Database: DatabaseConfig{
			Path: defaultDatabasePath,
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
	return nil
}
