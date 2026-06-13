package collaborators

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/everscribe/cli/internal/testutil"
	"github.com/everscribe/cli/internal/types"
)

func TestRunList_HappyPath(t *testing.T) {
	var gotMethod, gotPath string
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_ = json.NewEncoder(w).Encode(types.ListCollaboratorsResponse{Collaborators: []types.Collaborator{
			{ID: "c1", Email: "alice@example.com", Username: "alice", Status: "active",
				Projects: []types.CollaboratorProject{{ID: "p1", Name: "dev", Role: "viewer"}}},
			{ID: "c2", Email: "bob@example.com", Status: "pending"},
		}})
	})

	var buf bytes.Buffer
	require.NoError(t, runList(t.Context(), &buf, "table"))

	require.Equal(t, http.MethodGet, gotMethod)
	require.Equal(t, "/v1/collaborators", gotPath)
	out := buf.String()
	require.Contains(t, out, "alice@example.com")
	require.Contains(t, out, "pending")
}

func TestRunAdd_SendsEmail(t *testing.T) {
	var gotMethod, gotBody string
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(types.AddCollaboratorResponse{Collaborator: types.Collaborator{
			ID: "c1", Email: "newperson@example.com", Status: "pending",
		}})
	})

	var buf bytes.Buffer
	require.NoError(t, runAdd(t.Context(), &buf, "newperson@example.com", "table"))

	require.Equal(t, http.MethodPost, gotMethod)
	require.Contains(t, gotBody, "newperson@example.com")
	require.Contains(t, buf.String(), "newperson@example.com")
}

func TestRunRemove_ResolvesEmailToID(t *testing.T) {
	var deletePath string
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(types.ListCollaboratorsResponse{Collaborators: []types.Collaborator{
				{ID: "c1", Email: "alice@example.com", Status: "active"},
			}})
		case http.MethodDelete:
			deletePath = r.URL.Path
			w.WriteHeader(http.StatusNoContent)
		}
	})

	var buf bytes.Buffer
	require.NoError(t, runRemove(t.Context(), &buf, "alice@example.com"))

	require.Equal(t, "/v1/collaborators/c1", deletePath)
	require.Contains(t, buf.String(), "Removed alice@example.com")
}

func TestRunRemove_NotACollaborator(t *testing.T) {
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(types.ListCollaboratorsResponse{})
	})

	err := runRemove(t.Context(), io.Discard, "nobody@example.com")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not one of your collaborators")
}
