// Command es is the Everscribe command-line interface.
//
// Install: go install github.com/everscribe/cli/cmd/es@latest
package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/everscribe/cli/internal/cmds"
)

// version is overridable at link time:
//
//	go install -ldflags "-X main.version=v0.1.0" ./cmd/es
var version = "dev"

func main() {
	info, ok := debug.ReadBuildInfo()
	if err := cmds.NewRoot(resolveVersion(version, info, ok)).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// resolveVersion prefers the link-time version, then the module version
// that go install records, then "dev".
func resolveVersion(linked string, info *debug.BuildInfo, ok bool) string {
	if linked != "dev" {
		return linked
	}
	if ok && info != nil && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return linked
}
