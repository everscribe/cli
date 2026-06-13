package projects

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

func TestRunShare_StrictResolvesCollaborator(t *testing.T) {
	var putPath, putBody string
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/collaborators":
			_ = json.NewEncoder(w).Encode(types.ListCollaboratorsResponse{Collaborators: []types.Collaborator{
				{ID: "c1", Email: "alice@example.com", UserID: "u_alice", Status: "active"},
			}})
		case r.Method == http.MethodPut:
			putPath = r.URL.Path
			b, _ := io.ReadAll(r.Body)
			putBody = string(b)
			w.WriteHeader(http.StatusNoContent)
		}
	})

	var buf bytes.Buffer
	require.NoError(t, runShare(t.Context(), &buf, "proj-1", "alice@example.com", "editor"))

	require.Equal(t, "/v1/projects/proj-1/members/u_alice", putPath)
	require.Contains(t, putBody, "editor")
	require.Contains(t, buf.String(), "Shared project proj-1 with alice@example.com as editor")
}

func TestRunShare_RejectsNonCollaborator(t *testing.T) {
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(types.ListCollaboratorsResponse{})
	})

	err := runShare(t.Context(), io.Discard, "proj-1", "stranger@example.com", "viewer")
	require.Error(t, err)
	require.Contains(t, err.Error(), "add them first")
}

func TestRunShare_RejectsPending(t *testing.T) {
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(types.ListCollaboratorsResponse{Collaborators: []types.Collaborator{
			{ID: "c1", Email: "bob@example.com", Status: "pending"},
		}})
	})

	err := runShare(t.Context(), io.Discard, "proj-1", "bob@example.com", "viewer")
	require.Error(t, err)
	require.Contains(t, err.Error(), "hasn't created an Everscribe account")
}

func TestRunShare_InvalidRole(t *testing.T) {
	err := runShare(t.Context(), io.Discard, "proj-1", "x@example.com", "superuser")
	require.Error(t, err)
	require.Contains(t, err.Error(), "role must be")
}

func TestRunUnshare_ResolvesFromMembers(t *testing.T) {
	var deletePath string
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(types.ListProjectMembersResponse{Members: []types.ProjectMemberView{
				{UserID: "u_alice", Email: "alice@example.com", ProjectRole: "viewer"},
			}})
		case http.MethodDelete:
			deletePath = r.URL.Path
			w.WriteHeader(http.StatusNoContent)
		}
	})

	var buf bytes.Buffer
	require.NoError(t, runUnshare(t.Context(), &buf, "proj-1", "alice@example.com"))
	require.Equal(t, "/v1/projects/proj-1/members/u_alice", deletePath)
}

func TestRunMembers_Table(t *testing.T) {
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(types.ListProjectMembersResponse{Members: []types.ProjectMemberView{
			{UserID: "u1", Username: "owner", Email: "owner@example.com", AccountRole: "owner", ProjectRole: "admin"},
			{UserID: "u2", Username: "alice", Email: "alice@example.com", ProjectRole: "viewer"},
		}})
	})

	var buf bytes.Buffer
	require.NoError(t, runMembers(t.Context(), &buf, "proj-1", "table"))
	out := buf.String()
	require.Contains(t, out, "owner@example.com")
	require.Contains(t, out, "owner")
	require.Contains(t, out, "viewer")
}
