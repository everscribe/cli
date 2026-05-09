// Command es is the Everscribe command-line interface.
//
// Install: go install github.com/everscribe/cli/cmd/es@latest
package main

import (
	"fmt"
	"os"

	"github.com/everscribe/cli/internal/cmds"
)

// version is overridable at link time:
//
//	go install -ldflags "-X main.version=v0.1.0" ./cmd/es
var version = "dev"

func main() {
	if err := cmds.NewRoot(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
