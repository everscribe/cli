package events

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/everscribe/cli/internal/config"
	"github.com/everscribe/cli/internal/testutil"
	"github.com/everscribe/cli/internal/types"
)

// syncBuffer is a goroutine-safe bytes.Buffer. The watch test runs
// runWatch on a goroutine while the main test goroutine reads stdout
// — without locking, the race detector trips on concurrent Write/String.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func newSampleEvent(id, action string) types.Event {
	return types.Event{
		ID:         id,
		TenantID:   "acme-co",
		OccurredAt: time.Now().Add(-time.Minute),
		Action:     action,
		Actor:      json.RawMessage(`{"type":"user","id":"u_1","display_name":"alice"}`),
		Target:     json.RawMessage(`{"type":"session","id":"s_99"}`),
		Result:     json.RawMessage(`{"status":"ok"}`),
	}
}

func TestRunList_HappyPath(t *testing.T) {
	var gotPath, gotQuery, gotAuth string
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery, gotAuth = r.URL.Path, r.URL.RawQuery, r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(types.ListEventsResponse{
			Events: []types.Event{newSampleEvent("evt-1", "user.login")},
		})
	})

	var buf bytes.Buffer
	require.NoError(t, runList(context.Background(), &buf, "proj-001", filterFlags{}, 50, false, "table"))

	require.Equal(t, "/v1/projects/proj-001/events", gotPath)
	require.Contains(t, gotQuery, "limit=50")
	require.Equal(t, "Bearer pat_test_token", gotAuth)
	require.Contains(t, buf.String(), "user.login")
}

func TestRunList_FiltersAreSentAsQueryParams(t *testing.T) {
	var gotQuery string
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(types.ListEventsResponse{})
	})

	ff := filterFlags{
		action:     "user.login",
		actor:      "u_1",
		actorType:  "user",
		targetType: "session",
		tenant:     "acme-co",
	}
	require.NoError(t, runList(context.Background(), io.Discard, "proj-001", ff, 25, false, "table"))

	for _, want := range []string{"action=user.login", "actor=u_1", "actor_type=user", "target_type=session", "tenant_id=acme-co", "limit=25"} {
		require.Containsf(t, gotQuery, want, "missing query param %q in %q", want, gotQuery)
	}
}

func TestRunList_AllPaginatesUntilNoCursor(t *testing.T) {
	calls := 0
	cursors := []string{"cur-1", "cur-2", ""}
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		i := calls
		calls++
		_ = json.NewEncoder(w).Encode(types.ListEventsResponse{
			Events:     []types.Event{newSampleEvent("evt-"+cursors[i]+"a", "x.y")},
			NextCursor: cursors[i],
		})
	})

	var buf bytes.Buffer
	require.NoError(t, runList(context.Background(), &buf, "proj-001", filterFlags{}, 1, true, "json"))

	require.Equal(t, 3, calls, "should follow cursor until empty")
	// JSON output should be a single bare array containing all 3 events.
	out := strings.TrimSpace(buf.String())
	require.True(t, strings.HasPrefix(out, "["))
	require.Equal(t, 3, strings.Count(out, `"action": "x.y"`))
}

func TestRunList_PaginationHintInTableOnlyWithoutAll(t *testing.T) {
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(types.ListEventsResponse{
			Events:     []types.Event{newSampleEvent("evt-1", "a.b")},
			NextCursor: "cur-1",
		})
	})

	var tableBuf, jsonBuf bytes.Buffer
	require.NoError(t, runList(context.Background(), &tableBuf, "proj-001", filterFlags{}, 50, false, "table"))
	require.NoError(t, runList(context.Background(), &jsonBuf, "proj-001", filterFlags{}, 50, false, "json"))

	require.Contains(t, tableBuf.String(), "More events available", "table should hint at more pages")
	require.NotContains(t, jsonBuf.String(), "More events available", "json must not pollute machine output with hints")
}

func TestRunList_BadTimeFlagRejected(t *testing.T) {
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("API should not be called when --since is invalid")
	})

	err := runList(context.Background(), io.Discard, "proj-001", filterFlags{since: "not-a-time"}, 50, false, "table")
	require.Error(t, err)
	require.Contains(t, err.Error(), "--since")
}

func TestRunList_NotLoggedIn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	err := runList(context.Background(), io.Discard, "proj-001", filterFlags{}, 50, false, "table")
	require.ErrorIs(t, err, config.ErrNotLoggedIn)
}

func TestRunDescribe_HappyPathYAML(t *testing.T) {
	var gotPath string
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(types.EventResponse{Event: newSampleEvent("evt-1", "user.login")})
	})

	var buf bytes.Buffer
	require.NoError(t, runDescribe(context.Background(), &buf, "proj-001", "evt-1", "yaml"))

	require.Equal(t, "/v1/projects/proj-001/events/evt-1", gotPath)
	out := buf.String()
	require.Contains(t, out, "action: user.login")
	require.Contains(t, out, "display_name: alice", "actor sub-fields should be decoded inline")
}

func TestRunDescribe_RejectsTableFormat(t *testing.T) {
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {})
	err := runDescribe(context.Background(), io.Discard, "proj-001", "evt-1", "table")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not supported")
}

func TestRunDescribe_JSONFormat(t *testing.T) {
	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(types.EventResponse{Event: newSampleEvent("evt-1", "user.login")})
	})

	var buf bytes.Buffer
	require.NoError(t, runDescribe(context.Background(), &buf, "proj-001", "evt-1", "json"))
	out := buf.String()
	require.Contains(t, out, `"action": "user.login"`)
	require.Contains(t, out, `"display_name": "alice"`, "nested actor fields should appear")
}

func TestRunDiff_HappyPath(t *testing.T) {
	e := newSampleEvent("evt-1", "project.update")
	e.Change = json.RawMessage(`{"before":{"name":"old"},"after":{"name":"new"}}`)

	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(types.EventResponse{Event: e})
	})

	var buf bytes.Buffer
	require.NoError(t, runDiff(context.Background(), &buf, "proj-001", "evt-1"))

	out := buf.String()
	require.Contains(t, out, "--- before")
	require.Contains(t, out, "+++ after")
	require.Contains(t, out, `-  "name": "old"`)
	require.Contains(t, out, `+  "name": "new"`)
}

func TestRunDiff_NoChangeFieldErrorsClearly(t *testing.T) {
	e := newSampleEvent("evt-1", "user.login") // no Change

	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(types.EventResponse{Event: e})
	})

	err := runDiff(context.Background(), io.Discard, "proj-001", "evt-1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "no change record")
	require.Contains(t, err.Error(), "evt-1")
}

func TestRunDiff_NotLoggedIn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	err := runDiff(context.Background(), io.Discard, "proj-001", "evt-1")
	require.ErrorIs(t, err, config.ErrNotLoggedIn)
}

// TestRunWatch_TickReceivesNewEvents drives the watch loop through a
// single tick by canceling the context after the first poll completes.
// Verifies dedupe semantics and that the second batch's row appears
// in stdout.
func TestRunWatch_TickReceivesNewEvents(t *testing.T) {
	calls := 0
	first := []types.Event{newSampleEvent("evt-1", "first.event")}
	second := []types.Event{
		newSampleEvent("evt-1", "first.event"), // duplicate, must be deduped
		newSampleEvent("evt-2", "second.event"),
	}
	pollDone := make(chan struct{}, 4)

	testutil.SetupSession(t, func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			select {
			case pollDone <- struct{}{}:
			default:
			}
		}()
		i := calls
		calls++
		var batch []types.Event
		switch i {
		case 0:
			batch = first
		case 1:
			batch = second
		default:
			batch = nil
		}
		_ = json.NewEncoder(w).Encode(types.ListEventsResponse{Events: batch})
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	stdout := &syncBuffer{}

	// Use a tiny interval so the test runs quickly.
	go func() {
		err := runWatch(ctx, stdout, "proj-001", filterFlags{}, 25*time.Millisecond, "table")
		require.True(t, errors.Is(err, context.Canceled), "expected context.Canceled, got %v", err)
	}()

	// Wait for at least 2 polls to land before checking output.
	for i := 0; i < 2; i++ {
		select {
		case <-pollDone:
		case <-time.After(2 * time.Second):
			t.Fatalf("only %d polls completed", i)
		}
	}
	// Give the second batch a moment to render before we cancel.
	time.Sleep(50 * time.Millisecond)
	cancel()

	out := stdout.String()
	require.Contains(t, out, "first.event", "first batch should print")
	require.Contains(t, out, "second.event", "new event in second batch should print")
	require.Equal(t, 1, strings.Count(out, "first.event"), "duplicate should be deduped")
}
