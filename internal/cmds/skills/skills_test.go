package skills

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/everscribe/cli/internal/types"
)

// mockCDN bundles a httptest.Server with a mutable catalog + body —
// tests configure both via the returned pointers AFTER the server is
// up, so they can plug the server's URL into the catalog's Source
// fields without a chicken-and-egg dance.
type mockCDN struct {
	*httptest.Server
	catalog *types.SkillCatalog
	body    *string
}

func newMockCDN(t *testing.T) *mockCDN {
	t.Helper()
	cdn := &mockCDN{
		catalog: &types.SkillCatalog{Schema: 1},
		body:    new(string),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/manifests/skills.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cdn.catalog)
	})
	mux.HandleFunc("/manifests/setup.md", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/markdown")
		_, _ = w.Write([]byte(*cdn.body))
	})
	cdn.Server = httptest.NewServer(mux)
	t.Cleanup(cdn.Server.Close)
	t.Setenv(manifestsOverrideEnv, cdn.URL)
	return cdn
}

// withSetupSkill sets up the canonical "setup" skill in the catalog
// + body. Returns the same cdn so callers can chain.
func (m *mockCDN) withSetupSkill(body string) *mockCDN {
	*m.body = body
	m.catalog.Skills = []types.Skill{{
		Name:                "setup",
		Description:         "Set up Everscribe in your codebase",
		Version:             "1.0.0",
		Source:              m.URL + "/manifests/setup.md",
		ClaudeCodeSkillName: "everscribe-setup",
	}}
	return m
}

// --- list ---

func TestRunList_HappyPath_Table(t *testing.T) {
	newMockCDN(t).withSetupSkill("")

	var buf bytes.Buffer
	require.NoError(t, runList(t.Context(), &buf, "table"))

	out := buf.String()
	require.Contains(t, out, "setup")
	require.Contains(t, out, "1.0.0")
	require.Contains(t, out, "Set up Everscribe")
}

func TestRunList_HappyPath_JSON(t *testing.T) {
	newMockCDN(t).withSetupSkill("")

	var buf bytes.Buffer
	require.NoError(t, runList(t.Context(), &buf, "json"))

	// Bare array, not envelope-wrapped — matches `es projects list` convention.
	var got []types.Skill
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	require.Len(t, got, 1)
	require.Equal(t, "setup", got[0].Name)
}

func TestRunList_EmptyCatalog(t *testing.T) {
	newMockCDN(t)

	var buf bytes.Buffer
	require.NoError(t, runList(t.Context(), &buf, "table"))
	require.Contains(t, buf.String(), "No skills.")
}

func TestRunList_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	t.Setenv(manifestsOverrideEnv, srv.URL)

	err := runList(t.Context(), io.Discard, "table")
	require.Error(t, err)
	require.Contains(t, err.Error(), "HTTP 500")
}

func TestRunList_InvalidFormat(t *testing.T) {
	newMockCDN(t).withSetupSkill("")

	err := runList(t.Context(), io.Discard, "xml")
	require.Error(t, err)
}

// --- download ---

func TestRunDownload_HappyPath(t *testing.T) {
	const body = "---\nname: everscribe-setup\n---\n# Everscribe setup skill\n"
	newMockCDN(t).withSetupSkill(body)
	skillsRoot := t.TempDir()

	var buf bytes.Buffer
	require.NoError(t, runDownload(t.Context(), &buf, "setup", skillsRoot, false))

	wrote := filepath.Join(skillsRoot, "everscribe-setup", "SKILL.md")
	got, err := os.ReadFile(wrote)
	require.NoError(t, err)
	require.Equal(t, body, string(got))

	require.Contains(t, buf.String(), wrote)
	require.Contains(t, buf.String(), "v1.0.0")
	require.Contains(t, buf.String(), "/everscribe-setup")
}

func TestRunDownload_UnknownSkill(t *testing.T) {
	newMockCDN(t).withSetupSkill("")

	err := runDownload(t.Context(), io.Discard, "does-not-exist", t.TempDir(), false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestRunDownload_RefusesOverwriteWithoutForce(t *testing.T) {
	const body = "# v2\n"
	newMockCDN(t).withSetupSkill(body)
	skillsRoot := t.TempDir()
	existing := filepath.Join(skillsRoot, "everscribe-setup", "SKILL.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(existing), 0o755))
	require.NoError(t, os.WriteFile(existing, []byte("# pre-existing\n"), 0o644))

	err := runDownload(t.Context(), io.Discard, "setup", skillsRoot, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "--force")

	got, _ := os.ReadFile(existing)
	require.Equal(t, "# pre-existing\n", string(got))
}

func TestRunDownload_OverwritesWithForce(t *testing.T) {
	const body = "# v2\n"
	newMockCDN(t).withSetupSkill(body)
	skillsRoot := t.TempDir()
	existing := filepath.Join(skillsRoot, "everscribe-setup", "SKILL.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(existing), 0o755))
	require.NoError(t, os.WriteFile(existing, []byte("# pre-existing\n"), 0o644))

	require.NoError(t, runDownload(t.Context(), io.Discard, "setup", skillsRoot, true))

	got, _ := os.ReadFile(existing)
	require.Equal(t, body, string(got))
}

func TestRunDownload_EnvDirOverride(t *testing.T) {
	const body = "# from env\n"
	newMockCDN(t).withSetupSkill(body)
	skillsRoot := t.TempDir()
	t.Setenv(DefaultSkillsDirEnv, skillsRoot)

	// No --dir flag; env var should win.
	require.NoError(t, runDownload(t.Context(), io.Discard, "setup", "", false))

	got, err := os.ReadFile(filepath.Join(skillsRoot, "everscribe-setup", "SKILL.md"))
	require.NoError(t, err)
	require.Equal(t, body, string(got))
}

// --- fetcher edge case ---

func TestFetcher_SkillBody_EmptyURL(t *testing.T) {
	f := NewFetcher()
	_, err := f.SkillBody(t.Context(), "")
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "empty"))
}
