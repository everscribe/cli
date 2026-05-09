package cobrax

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestRequireArgs_Exact(t *testing.T) {
	v := RequireArgs("id")
	require.NoError(t, v(&cobra.Command{}, []string{"abc"}))
}

func TestRequireArgs_MissingNamesTheArg(t *testing.T) {
	v := RequireArgs("event-id")
	err := v(&cobra.Command{}, nil)
	require.Error(t, err)
	require.Equal(t, "missing required argument: <event-id>", err.Error())
}

func TestRequireArgs_MissingNamesTheNextArg(t *testing.T) {
	// Two args declared, only one supplied — error should name the
	// SECOND arg as the missing one, not the first.
	v := RequireArgs("project-id", "event-id")
	err := v(&cobra.Command{}, []string{"proj-1"})
	require.Error(t, err)
	require.Equal(t, "missing required argument: <event-id>", err.Error())
}

func TestRequireArgs_TooMany(t *testing.T) {
	v := RequireArgs("id")
	err := v(&cobra.Command{}, []string{"a", "b", "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected extra argument")
	require.Contains(t, err.Error(), "[b c]")
}

func TestRequireArgs_NoArgsAndNoNames(t *testing.T) {
	v := RequireArgs()
	require.NoError(t, v(&cobra.Command{}, nil))
}
