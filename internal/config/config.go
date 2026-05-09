// Package config persists the CLI's authenticated session to
// ~/.config/everscribe/pat.json.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PAT is the on-disk representation of an authenticated CLI session.
type PAT struct {
	Token     string    `json:"token"`
	PATID     string    `json:"pat_id"`
	UserID    string    `json:"user_id"`
	UserEmail string    `json:"user_email"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// ErrNotLoggedIn is returned by Load when no session file exists.
var ErrNotLoggedIn = errors.New("not logged in: run `es auth login`")

// Path returns the absolute path to the on-disk PAT file.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home dir: %w", err)
	}
	return filepath.Join(home, ".config", "everscribe", "pat.json"), nil
}

// Load reads the on-disk PAT. Returns ErrNotLoggedIn if the file is absent.
func Load() (*PAT, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotLoggedIn
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", p, err)
	}
	var pat PAT
	if err := json.Unmarshal(data, &pat); err != nil {
		return nil, fmt.Errorf("parse %s: %w", p, err)
	}
	if pat.Token == "" {
		return nil, ErrNotLoggedIn
	}
	return &pat, nil
}

// Save writes pat to disk, creating the parent directory with mode 0700 and
// the file with mode 0600.
func Save(pat *PAT) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(pat, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal pat: %w", err)
	}
	if err := os.WriteFile(p, data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", p, err)
	}
	return nil
}

// Delete removes the on-disk PAT. No-op if the file is already absent.
func Delete() error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove %s: %w", p, err)
	}
	return nil
}
