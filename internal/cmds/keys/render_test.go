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

// TestRenderKeys covers list rendering across formats and edge inputs
// (empty list, never-used key). Cases vary in input fixture and the
// expected substrings; the structural one-off (json starts with `[`)
// is asserted in extraCheck.
func TestRenderKeys(t *testing.T) {
	neverUsed := sampleKey()
	neverUsed.LastUsedAt = time.Time{}

	cases := []struct {
		name       string
		format     string
		input      []types.APIKey
		wantSub    []string
		extraCheck func(t *testing.T, out string)
	}{
		{
			name:    "table has expected columns",
			format:  "table",
			input:   []types.APIKey{sampleKey()},
			wantSub: []string{"ID", "NAME", "PREFIX", "LAST USED", "AGE", "key-001", "production", "evs_a1b2c3d4", "ago"},
		},
		{
			name:    "never-used key shows 'never'",
			format:  "table",
			input:   []types.APIKey{neverUsed},
			wantSub: []string{"never"},
		},
		{
			name:    "empty table prints message",
			format:  "table",
			input:   nil,
			wantSub: []string{"No API keys."},
		},
		{
			name:   "json is bare array",
			format: "json",
			input:  []types.APIKey{sampleKey()},
			extraCheck: func(t *testing.T, out string) {
				trimmed := strings.TrimSpace(out)
				require.True(t, strings.HasPrefix(trimmed, "["), "list JSON should be a bare array, got:\n%s", out)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, renderKeys(&buf, tc.format, tc.input))
			out := buf.String()
			for _, want := range tc.wantSub {
				require.Containsf(t, out, want, "missing %q", want)
			}
			if tc.extraCheck != nil {
				tc.extraCheck(t, out)
			}
		})
	}
}

// TestRenderCreatedKey covers the post-create render: same response
// shape across formats, but each format must surface both the
// metadata and the one-time plaintext.
func TestRenderCreatedKey(t *testing.T) {
	resp := &types.CreateAPIKeyResponse{
		Key:       sampleKey(),
		Plaintext: "evs_a1b2c3d4e5f67890abcdef1234567890fedcba",
	}

	cases := []struct {
		name    string
		format  string
		wantSub []string
	}{
		{
			name:    "table shows banner and plaintext",
			format:  "table",
			wantSub: []string{"key-001", "evs_a1b2c3d4e5f67890abcdef1234567890fedcba", "will not be shown again"},
		},
		{
			name:    "json includes plaintext",
			format:  "json",
			wantSub: []string{`"plaintext": "evs_a1b2c3d4e5f67890abcdef1234567890fedcba"`, `"id": "key-001"`},
		},
		{
			name:    "yaml includes plaintext",
			format:  "yaml",
			wantSub: []string{"plaintext: evs_a1b2c3d4e5f67890abcdef1234567890fedcba"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, renderCreatedKey(&buf, tc.format, resp))
			out := buf.String()
			for _, want := range tc.wantSub {
				require.Containsf(t, out, want, "missing %q", want)
			}
		})
	}
}

func TestRenderKeys_BadFormatRejected(t *testing.T) {
	err := renderKeys(&bytes.Buffer{}, "xml", []types.APIKey{sampleKey()})
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid")
}
