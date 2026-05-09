package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/everscribe/cli/internal/config"
	"github.com/everscribe/cli/internal/testutil"
	"github.com/everscribe/cli/internal/types"
)

func sampleProject() types.Project {
	return types.Project{
		ID:        "proj-001",
		UserID:    "u_test",
		Name:      "ingest-pipeline",
		CreatedAt: time.Now().Add(-2 * time.Hour),
		UpdatedAt: time.Now(),
	}
}

func TestRunList_HappyPath(t *testing.T) {
	var gotMethod, gotPath, gotAuth string
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotAuth = r.Method, r.URL.Path, r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(types.ListProjectsResponse{Projects: []types.Project{sampleProject()}})
	})

	var buf bytes.Buffer
	require.NoError(t, runList(context.Background(), &buf, "table"))

	require.Equal(t, http.MethodGet, gotMethod)
	require.Equal(t, "/v1/projects", gotPath)
	require.Equal(t, "Bearer pat_test_token", gotAuth)
	require.Contains(t, buf.String(), "ingest-pipeline")
}

func TestRunList_NotLoggedIn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	err := runList(context.Background(), io.Discard, "table")
	require.ErrorIs(t, err, config.ErrNotLoggedIn)
}

func TestRunGet_HappyPath(t *testing.T) {
	var gotPath string
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(types.ProjectResponse{Project: sampleProject()})
	})

	var buf bytes.Buffer
	require.NoError(t, runGet(context.Background(), &buf, "proj-001", "yaml"))

	require.Equal(t, "/v1/projects/proj-001", gotPath)
	require.Contains(t, buf.String(), "name: ingest-pipeline")
}

func TestRunCreate_SendsBodyAndPrintsResult(t *testing.T) {
	var gotMethod, gotPath, gotBody string
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		_ = json.NewEncoder(w).Encode(types.ProjectResponse{Project: sampleProject()})
	})

	var buf bytes.Buffer
	require.NoError(t, runCreate(context.Background(), &buf, "ingest-pipeline", "json"))

	require.Equal(t, http.MethodPost, gotMethod)
	require.Equal(t, "/v1/projects", gotPath)
	require.JSONEq(t, `{"name":"ingest-pipeline"}`, gotBody)
	require.Contains(t, buf.String(), `"id": "proj-001"`)
}

func TestRunUpdate_SendsPutAndBody(t *testing.T) {
	var gotMethod, gotPath, gotBody string
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		updated := sampleProject()
		updated.Name = "renamed"
		_ = json.NewEncoder(w).Encode(types.ProjectResponse{Project: updated})
	})

	var buf bytes.Buffer
	require.NoError(t, runUpdate(context.Background(), &buf, "proj-001", "renamed", "table"))

	require.Equal(t, http.MethodPut, gotMethod)
	require.Equal(t, "/v1/projects/proj-001", gotPath)
	require.JSONEq(t, `{"name":"renamed"}`, gotBody)
	require.Contains(t, buf.String(), "renamed")
}

func TestRunDelete_SendsDelete(t *testing.T) {
	var gotMethod, gotPath string
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	var buf bytes.Buffer
	require.NoError(t, runDelete(context.Background(), &buf, "proj-001"))

	require.Equal(t, http.MethodDelete, gotMethod)
	require.Equal(t, "/v1/projects/proj-001", gotPath)
	require.Contains(t, buf.String(), "Project proj-001 deleted")
}

// TestRunCreate_PropagatesAPIError verifies that server errors flow
// through as APIError; commands rely on this for exit codes and the
// root cobra error printer for the user-visible message.
func TestRunCreate_PropagatesAPIError(t *testing.T) {
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "project name already taken", http.StatusConflict)
	})

	err := runCreate(context.Background(), io.Discard, "duplicate", "table")
	require.Error(t, err)
	require.Contains(t, err.Error(), "already taken")
}

// Sanity that the cobra command tree wires everything up — exercising
// New/CreateCmd/etc. by invoking the parent and walking subcommands.
func TestNewCmd_HasAllSubcommands(t *testing.T) {
	cmd := NewCmd()
	subs := map[string]bool{}
	for _, c := range cmd.Commands() {
		subs[c.Name()] = true
	}
	for _, want := range []string{"list", "get", "create", "update", "delete", "use", "current"} {
		require.Truef(t, subs[want], "missing subcommand %q", want)
	}
}

func TestRunUse_ValidatesAndSavesDefault(t *testing.T) {
	var gotPath string
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(types.ProjectResponse{Project: sampleProject()})
	})

	var buf bytes.Buffer
	require.NoError(t, runUse(context.Background(), &buf, "proj-001"))
	require.Equal(t, "/v1/projects/proj-001", gotPath, "should validate via GetProject before persisting")

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.Equal(t, "proj-001", cfg.DefaultProjectID)

	require.Contains(t, buf.String(), "ingest-pipeline", "confirmation message should include the project name")
}

func TestRunUse_RejectsUnknownProject(t *testing.T) {
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "project not found", http.StatusNotFound)
	})

	err := runUse(context.Background(), io.Discard, "no-such-id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.Empty(t, cfg.DefaultProjectID, "must not persist a typo'd project ID")
}

func TestRunCurrent_PrintsSavedDefault(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	require.NoError(t, config.SaveConfig(&config.CLIConfig{DefaultProjectID: "proj-001"}))

	var buf bytes.Buffer
	require.NoError(t, runCurrent(&buf))
	require.Equal(t, "proj-001\n", buf.String())
}

func TestRunCurrent_NoneSet(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	var buf bytes.Buffer
	require.NoError(t, runCurrent(&buf))
	require.Contains(t, buf.String(), "No default project set")
}

func TestRunList_BadFormatRejected(t *testing.T) {
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(types.ListProjectsResponse{})
	})

	err := runList(context.Background(), io.Discard, "xml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid")
}
