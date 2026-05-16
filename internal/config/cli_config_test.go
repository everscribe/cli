package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCLIConfig_RoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	require.NoError(t, SaveConfig(&CLIConfig{DefaultProjectID: "proj-001"}))
	got, err := LoadConfig()
	require.NoError(t, err)
	require.Equal(t, "proj-001", got.DefaultProjectID)
}

func TestLoadConfig_MissingFileIsEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	got, err := LoadConfig()
	require.NoError(t, err)
	require.Empty(t, got.DefaultProjectID, "fresh install should produce an empty config without error")
}

func TestSaveConfig_FilePerms(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	require.NoError(t, SaveConfig(&CLIConfig{DefaultProjectID: "p"}))

	p, err := ConfigPath()
	require.NoError(t, err)
	info, err := os.Stat(p)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestResolveProjectID_FlagWins(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	require.NoError(t, SaveConfig(&CLIConfig{DefaultProjectID: "from-config"}))

	got, err := ResolveProjectID("from-flag")
	require.NoError(t, err)
	require.Equal(t, "from-flag", got, "explicit flag must override saved default")
}

func TestResolveProjectID_FallsBackToConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	require.NoError(t, SaveConfig(&CLIConfig{DefaultProjectID: "from-config"}))

	got, err := ResolveProjectID("")
	require.NoError(t, err)
	require.Equal(t, "from-config", got)
}

func TestResolveProjectID_NeitherSet(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := ResolveProjectID("")
	require.Error(t, err)
	require.ErrorContains(t, err, "no project specified")
	require.ErrorContains(t, err, "es projects use")
}
