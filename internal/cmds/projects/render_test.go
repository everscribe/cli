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

func TestRenderProjects_TableHasExpectedColumns(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, renderProjects(&buf, "table", fixtureProjects))

	out := buf.String()
	for _, want := range []string{"ID", "NAME", "AGE", "proj-001", "ingest-pipeline", "proj-002", "alpha"} {
		require.Containsf(t, out, want, "table missing %q", want)
	}
	// Age formatting is delegated to output.Age — sanity check we got
	// a relative duration ("d" or "h" or "m"), not a full timestamp.
	require.Truef(t, strings.ContainsAny(out, "dhms"), "expected relative age in output:\n%s", out)
}

func TestRenderProjects_EmptyTablePrintsMessage(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, renderProjects(&buf, "table", nil))
	require.Contains(t, buf.String(), "No projects.")
}

func TestRenderProjects_JSONIsBareArray(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, renderProjects(&buf, "json", fixtureProjects))
	out := buf.String()
	require.True(t, strings.HasPrefix(strings.TrimSpace(out), "["),
		"list JSON should be a bare array, got:\n%s", out)
	require.NotContains(t, out, `"projects"`, "no envelope expected")
	require.Contains(t, out, `"name": "ingest-pipeline"`)
}

func TestRenderProjects_YAML(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, renderProjects(&buf, "yaml", fixtureProjects))
	require.Contains(t, buf.String(), "name: ingest-pipeline")
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
	require.Contains(t, err.Error(), "invalid")
}
