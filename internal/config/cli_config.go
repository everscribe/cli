package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// CLIConfig holds non-sensitive CLI preferences. Persisted alongside
// pat.json at ~/.config/everscribe/config.json (mode 0600). Kept in
// a separate file so `es auth logout` doesn't blow away preferences
// like the user's default project.
type CLIConfig struct {
	DefaultProjectID string `json:"default_project_id,omitempty"`
}

const configFileName = "config.json"

// ConfigPath returns the absolute path to the on-disk preferences file.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home dir: %w", err)
	}
	return filepath.Join(home, ".config", "everscribe", configFileName), nil
}

// LoadConfig reads the preferences file. A missing file is not an
// error — first-run UX shouldn't require explicit init.
func LoadConfig() (*CLIConfig, error) {
	p, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return &CLIConfig{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", p, err)
	}
	var c CLIConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", p, err)
	}
	return &c, nil
}

// SaveConfig writes c to disk, creating the parent directory with mode
// 0700 and the file with mode 0600.
func SaveConfig(c *CLIConfig) error {
	p, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(p, data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", p, err)
	}
	return nil
}

// ResolveProjectID returns the project ID a command should use. The
// flag value (if non-empty) takes precedence; otherwise the saved
// default. Returns a friendly error citing the right command if
// neither is set.
func ResolveProjectID(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	c, err := LoadConfig()
	if err != nil {
		return "", err
	}
	if c.DefaultProjectID == "" {
		return "", errors.New("No project specified. Pass --project <id> or set a default with `es projects use <id>`")
	}
	return c.DefaultProjectID, nil
}
