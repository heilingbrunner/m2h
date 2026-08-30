package main

import "github.com/heilingbrunner/m2h/cmd"

// version is set via -ldflags "-X main.version=…" at build time; defaults to
// "dev" for plain `go build` / `make` invocations.
var version = "dev"

func main() {
	cmd.Execute(version)
}
