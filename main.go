package main

import (
	"runtime/debug"

	"github.com/heilingbrunner/m2h/cmd"
)

// version is stamped via -ldflags "-X main.version=…" by the makefile and by
// goreleaser. For `go install github.com/heilingbrunner/m2h@version` builds no
// ldflags are applied, so it stays empty and is resolved from the module's
// embedded build info instead (see resolveVersion).
var version = ""

func main() {
	cmd.Execute(resolveVersion())
}

// resolveVersion returns the linker-stamped version when present, otherwise the
// module version Go embeds into `go install module@version` binaries, and
// finally "dev" for a plain `go build` from a source tree.
func resolveVersion() string {
	if version != "" {
		return version
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}

	return "dev"
}
