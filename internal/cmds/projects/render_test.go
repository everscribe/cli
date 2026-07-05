package projects

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/everscribe/cli/internal/types"
)

var fixtureProjects = []types.Project{
	{
		ID:        "proj-001",
		UserID:    "u_1",
		Name:      "ingest-pipeline",
		CreatedAt: time.Now().Add(-3 * 24 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	},
	{
		ID:        "proj-002",
		UserID:    "u_1",
		Name:      "alpha",
		CreatedAt: time.Now().Add(-90 * time.Minute),
		UpdatedAt: time.Now().Add(-90 * time.Minute),
	},
}

// TestRenderProjects covers list rendering across formats and edge
// inputs. extraCheck runs format-specific structural assertions that
// don't fit the substring pattern.
func TestRenderProjects(t *testing.T) {
	cases := []struct {
		name       string
		format     string
		input      []types.Project
		wantSub    []string
		wantNotSub []string
		extraCheck func(t *testing.T, out string)
	}{
		{
			name:    "table has expected columns",
			format:  "table",
			input:   fixtureProjects,
			wantSub: []string{"ID", "NAME", "AGE", "proj-001", "ingest-pipeline", "proj-002", "alpha"},
			extraCheck: func(t *testing.T, out string) {
				// Age formatting is delegated to output.Age - sanity check
				// we got a relative duration ("d" or "h" or "m"), not a
				// full timestamp.
				require.Truef(t, strings.ContainsAny(out, "dhms"), "expected relative age in output:\n%s", out)
			},
		},
		{
			name:    "empty table prints message",
			format:  "table",
			input:   nil,
			wantSub: []string{"No projects."},
		},
		{
			name:       "json is bare array",
			format:     "json",
			input:      fixtureProjects,
			wantSub:    []string{`"name": "ingest-pipeline"`},
			wantNotSub: []string{`"projects"`},
			extraCheck: func(t *testing.T, out string) {
				require.True(t, strings.HasPrefix(strings.TrimSpace(out), "["),
					"list JSON should be a bare array, got:\n%s", out)
			},
		},
		{
			name:    "yaml",
			format:  "yaml",
			input:   fixtureProjects,
			wantSub: []string{"name: ingest-pipeline"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, renderProjects(&buf, tc.format, tc.input))
			out := buf.String()
			for _, want := range tc.wantSub {
				require.Containsf(t, out, want, "missing %q", want)
			}
			for _, notWant := range tc.wantNotSub {
				require.NotContainsf(t, out, notWant, "should not contain %q", notWant)
			}
			if tc.extraCheck != nil {
				tc.extraCheck(t, out)
			}
		})
	}
}

func TestRenderProject_TableSingleRow(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, renderProject(&buf, "table", fixtureProjects[0]))
	out := buf.String()
	require.Contains(t, out, "proj-001")
	require.Contains(t, out, "ingest-pipeline")
	// One header row + one data row = 2 newlines minimum.
	require.GreaterOrEqual(t, strings.Count(out, "\n"), 2)
}

func TestRenderProject_JSONIsBareObject(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, renderProject(&buf, "json", fixtureProjects[0]))
	out := strings.TrimSpace(buf.String())
	require.True(t, strings.HasPrefix(out, "{"), "single project JSON should be an object, got:\n%s", out)
	require.NotContains(t, out, `"project":`, "no envelope expected")
}

func TestRenderProjects_RejectsBadFormat(t *testing.T) {
	err := renderProjects(&bytes.Buffer{}, "xml", fixtureProjects)
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid")
}
