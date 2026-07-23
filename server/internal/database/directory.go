package database

import (
	"fmt"
	"os"
)

func ensureDirectory(path string) error {
	if path == "." {
		return nil
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create database directory: %w", err)
	}
	return nil
}
