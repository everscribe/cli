package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
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
	if err := c.Do(context.Background(), http.MethodGet, "/v1/projects", q, nil, &out); err != nil {
		t.Fatalf("Do: %v", err)
	}

	if gotAuth != "Bearer pat_secret" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer pat_secret")
	}
	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/v1/projects" {
		t.Errorf("path = %q, want /v1/projects", gotPath)
	}
	if gotQuery != "limit=5" {
		t.Errorf("query = %q, want limit=5", gotQuery)
	}
	if gotAccept != "application/json" {
		t.Errorf("Accept = %q, want application/json", gotAccept)
	}
	if out.ID != "p_1" {
		t.Errorf("decoded ID = %q, want p_1", out.ID)
	}
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
	if err := c.Do(context.Background(), http.MethodPost, "/v1/projects", nil, in, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q", gotContentType)
	}
	if !strings.Contains(gotBody, `"name":"ingest-pipeline"`) {
		t.Errorf("body = %q, want JSON-encoded payload", gotBody)
	}
}

func TestDoNoTokenOmitsAuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := New("", WithBaseURL(srv.URL))
	if err := c.Do(context.Background(), http.MethodGet, "/health", nil, nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if gotAuth != "" {
		t.Errorf("Authorization = %q, want empty", gotAuth)
	}
}

func TestDoPlaintextErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "project name already taken", http.StatusConflict)
	}))
	t.Cleanup(srv.Close)

	c := New("pat_x", WithBaseURL(srv.URL))
	err := c.Do(context.Background(), http.MethodPost, "/v1/projects", nil, map[string]string{"name": "x"}, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusConflict {
		t.Errorf("StatusCode = %d, want 409", apiErr.StatusCode)
	}
	if !strings.Contains(apiErr.Message, "already taken") {
		t.Errorf("Message = %q, want plaintext body", apiErr.Message)
	}
}

func TestDoJSONErrorEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":"invalid_input","message":"name too long"}`))
	}))
	t.Cleanup(srv.Close)

	c := New("pat_x", WithBaseURL(srv.URL))
	err := c.Do(context.Background(), http.MethodPost, "/v1/projects", nil, map[string]string{}, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.Code != "invalid_input" || apiErr.Message != "name too long" {
		t.Errorf("apiErr = %+v, want code/message decoded from JSON", apiErr)
	}
}

func TestIsUnauthorizedAndNotFound(t *testing.T) {
	mk := func(status int) error {
		return &APIError{StatusCode: status, Message: http.StatusText(status)}
	}
	if !IsUnauthorized(mk(http.StatusUnauthorized)) {
		t.Error("IsUnauthorized 401 = false")
	}
	if IsUnauthorized(mk(http.StatusForbidden)) {
		t.Error("IsUnauthorized 403 = true")
	}
	if !IsNotFound(mk(http.StatusNotFound)) {
		t.Error("IsNotFound 404 = false")
	}
	if IsNotFound(errors.New("not an api error")) {
		t.Error("IsNotFound on non-APIError = true")
	}
}
