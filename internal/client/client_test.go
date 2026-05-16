package client

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDoSendsBearerAndDecodes(t *testing.T) {
	type response struct {
		ID string `json:"id"`
	}

	var gotAuth, gotMethod, gotPath, gotQuery, gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAccept = r.Header.Get("Accept")
		_ = json.NewEncoder(w).Encode(response{ID: "p_1"})
	}))
	t.Cleanup(srv.Close)

	c := New("pat_secret", WithBaseURL(srv.URL))
	var out response
	q := url.Values{"limit": {"5"}}
	require.NoError(t, c.Do(t.Context(), http.MethodGet, "/v1/projects", q, nil, &out))

	require.Equal(t, "Bearer pat_secret", gotAuth)
	require.Equal(t, http.MethodGet, gotMethod)
	require.Equal(t, "/v1/projects", gotPath)
	require.Equal(t, "limit=5", gotQuery)
	require.Equal(t, "application/json", gotAccept)
	require.Equal(t, "p_1", out.ID)
}

func TestDoSendsJSONBody(t *testing.T) {
	var gotBody, gotContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := New("pat_x", WithBaseURL(srv.URL))
	in := map[string]string{"name": "ingest-pipeline"}
	require.NoError(t, c.Do(t.Context(), http.MethodPost, "/v1/projects", nil, in, nil))
	require.Equal(t, "application/json", gotContentType)
	require.Contains(t, gotBody, `"name":"ingest-pipeline"`)
}

func TestDoNoTokenOmitsAuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := New("", WithBaseURL(srv.URL))
	require.NoError(t, c.Do(t.Context(), http.MethodGet, "/health", nil, nil, nil))
	require.Empty(t, gotAuth)
}

func TestDoPlaintextErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "project name already taken", http.StatusConflict)
	}))
	t.Cleanup(srv.Close)

	c := New("pat_x", WithBaseURL(srv.URL))
	err := c.Do(t.Context(), http.MethodPost, "/v1/projects", nil, map[string]string{"name": "x"}, nil)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, http.StatusConflict, apiErr.StatusCode)
	require.Contains(t, apiErr.Message, "already taken")
}

func TestDoJSONErrorEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":"invalid_input","message":"name too long"}`))
	}))
	t.Cleanup(srv.Close)

	c := New("pat_x", WithBaseURL(srv.URL))
	err := c.Do(t.Context(), http.MethodPost, "/v1/projects", nil, map[string]string{}, nil)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, "invalid_input", apiErr.Code)
	require.Equal(t, "name too long", apiErr.Message)
}

func TestIsUnauthorizedAndNotFound(t *testing.T) {
	mk := func(status int) error {
		return &APIError{StatusCode: status, Message: http.StatusText(status)}
	}
	require.True(t, IsUnauthorized(mk(http.StatusUnauthorized)))
	require.False(t, IsUnauthorized(mk(http.StatusForbidden)))
	require.True(t, IsNotFound(mk(http.StatusNotFound)))
	require.False(t, IsNotFound(errors.New("not an api error")))
}
