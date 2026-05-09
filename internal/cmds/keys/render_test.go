package keys

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/everscribe/cli/internal/types"
)

func sampleKey() types.APIKey {
	return types.APIKey{
		ID:         "key-001",
		ProjectID:  "proj-001",
		Name:       "production",
		Prefix:     "evs_a1b2c3d4",
		LastUsedAt: time.Now().Add(-5 * time.Minute),
		CreatedAt:  time.Now().Add(-3 * 24 * time.Hour),
	}
}

func TestRenderKeys_TableHasExpectedColumns(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, renderKeys(&buf, "table", []types.APIKey{sampleKey()}))

	out := buf.String()
	for _, want := range []string{"ID", "NAME", "PREFIX", "LAST USED", "AGE", "key-001", "production", "evs_a1b2c3d4"} {
		require.Containsf(t, out, want, "table missing %q", want)
	}
	require.Contains(t, out, "ago", "LAST USED column should render relative time")
}

func TestRenderKeys_NeverUsedShowsNever(t *testing.T) {
	k := sampleKey()
	k.LastUsedAt = time.Time{}

	var buf bytes.Buffer
	require.NoError(t, renderKeys(&buf, "table", []types.APIKey{k}))
	require.Contains(t, buf.String(), "never")
}

func TestRenderKeys_EmptyTable(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, renderKeys(&buf, "table", nil))
	require.Contains(t, buf.String(), "No API keys.")
}

func TestRenderKeys_JSONIsBareArray(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, renderKeys(&buf, "json", []types.APIKey{sampleKey()}))
	out := strings.TrimSpace(buf.String())
	require.True(t, strings.HasPrefix(out, "["), "list JSON should be a bare array, got:\n%s", out)
}

func TestRenderCreatedKey_TableShowsBannerAndPlaintext(t *testing.T) {
	resp := &types.CreateAPIKeyResponse{
		Key:       sampleKey(),
		Plaintext: "evs_a1b2c3d4e5f67890abcdef1234567890fedcba",
	}
	var buf bytes.Buffer
	require.NoError(t, renderCreatedKey(&buf, "table", resp))

	out := buf.String()
	require.Contains(t, out, "key-001", "metadata should be in the table")
	require.Contains(t, out, "evs_a1b2c3d4e5f67890abcdef1234567890fedcba",
		"plaintext must be printed for the user to copy")
	require.Contains(t, out, "will not be shown again",
		"banner should warn about one-time secret")
}

func TestRenderCreatedKey_JSONIncludesPlaintext(t *testing.T) {
	resp := &types.CreateAPIKeyResponse{
		Key:       sampleKey(),
		Plaintext: "evs_secret",
	}
	var buf bytes.Buffer
	require.NoError(t, renderCreatedKey(&buf, "json", resp))
	out := buf.String()
	require.Contains(t, out, `"plaintext": "evs_secret"`)
	require.Contains(t, out, `"id": "key-001"`)
}

func TestRenderCreatedKey_YAMLIncludesPlaintext(t *testing.T) {
	resp := &types.CreateAPIKeyResponse{
		Key:       sampleKey(),
		Plaintext: "evs_secret",
	}
	var buf bytes.Buffer
	require.NoError(t, renderCreatedKey(&buf, "yaml", resp))
	require.Contains(t, buf.String(), "plaintext: evs_secret")
}

func TestRenderKeys_BadFormatRejected(t *testing.T) {
	err := renderKeys(&bytes.Buffer{}, "xml", []types.APIKey{sampleKey()})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid")
}
