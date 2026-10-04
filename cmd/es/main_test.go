package main

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		linked string
		info   *debug.BuildInfo
		ok     bool
		want   string
	}{
		{name: "ldflags win", linked: "v0.2.0", info: &debug.BuildInfo{Main: debug.Module{Version: "v0.1.0"}}, ok: true, want: "v0.2.0"},
		{name: "go install module version", linked: "dev", info: &debug.BuildInfo{Main: debug.Module{Version: "v0.1.0"}}, ok: true, want: "v0.1.0"},
		{name: "local build devel", linked: "dev", info: &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, ok: true, want: "dev"},
		{name: "empty module version", linked: "dev", info: &debug.BuildInfo{}, ok: true, want: "dev"},
		{name: "no build info", linked: "dev", ok: false, want: "dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, resolveVersion(tt.linked, tt.info, tt.ok))
		})
	}
}
