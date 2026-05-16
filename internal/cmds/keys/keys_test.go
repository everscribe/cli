package keys

import (
	"bytes"
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

func freshKey() types.APIKey {
	return types.APIKey{
		ID:        "key-001",
		ProjectID: "proj-001",
		Name:      "production",
		Prefix:    "evs_a1b2c3d4",
		CreatedAt: time.Now(),
	}
}

func TestRunList_HappyPath(t *testing.T) {
	var gotMethod, gotPath, gotAuth string
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotAuth = r.Method, r.URL.Path, r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(types.ListAPIKeysResponse{Keys: []types.APIKey{freshKey()}})
	})

	var buf bytes.Buffer
	require.NoError(t, runList(t.Context(), &buf, "proj-001", "table"))

	require.Equal(t, http.MethodGet, gotMethod)
	require.Equal(t, "/v1/projects/proj-001/keys", gotPath)
	require.Equal(t, "Bearer pat_test_token", gotAuth)
	require.Contains(t, buf.String(), "production")
}

func TestRunList_NotLoggedIn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	err := runList(t.Context(), io.Discard, "proj-001", "table")
	require.ErrorIs(t, err, config.ErrNotLoggedIn)
}

func TestRunCreate_SendsBodyAndShowsPlaintext(t *testing.T) {
	var gotMethod, gotPath, gotBody string
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		_ = json.NewEncoder(w).Encode(types.CreateAPIKeyResponse{
			Key:       freshKey(),
			Plaintext: "evs_a1b2c3d4e5f67890abcdef1234567890fedcba",
		})
	})

	var buf bytes.Buffer
	require.NoError(t, runCreate(t.Context(), &buf, "proj-001", "production", "table"))

	require.Equal(t, http.MethodPost, gotMethod)
	require.Equal(t, "/v1/projects/proj-001/keys", gotPath)
	require.JSONEq(t, `{"name":"production"}`, gotBody)

	out := buf.String()
	require.Contains(t, out, "evs_a1b2c3d4e5f67890abcdef1234567890fedcba",
		"plaintext key must be surfaced")
	require.Contains(t, out, "will not be shown again")
}

func TestRunRevoke_SendsDelete(t *testing.T) {
	var gotMethod, gotPath string
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	var buf bytes.Buffer
	require.NoError(t, runRevoke(t.Context(), &buf, "proj-001", "key-001"))

	require.Equal(t, http.MethodDelete, gotMethod)
	require.Equal(t, "/v1/projects/proj-001/keys/key-001", gotPath)
	require.Contains(t, buf.String(), "Key key-001 revoked")
}

func TestRunCreate_PropagatesAPIError(t *testing.T) {
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "key name already taken", http.StatusConflict)
	})

	err := runCreate(t.Context(), io.Discard, "proj-001", "duplicate", "table")
	require.Error(t, err)
	require.ErrorContains(t, err, "already taken")
}

func TestNewCmd_HasAllSubcommands(t *testing.T) {
	cmd := NewCmd()
	subs := map[string]bool{}
	for _, c := range cmd.Commands() {
		subs[c.Name()] = true
	}
	for _, want := range []string{"list", "create", "revoke"} {
		require.Truef(t, subs[want], "missing subcommand %q", want)
	}
}
