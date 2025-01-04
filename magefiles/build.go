package main

import (
	"log/slog"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	mg.Deps(InstallDeps, InstallToolDeps)
	slog.Info("Building...")
	return Go("build", "-o", "bin/manager", "cmd/main.go")
}

func Run() error {
	mg.Deps(InstallDeps, InstallToolDeps)

	return Go("run", "./cmd/main.go")
}

func InstallToolDeps() error {
	return Go("mod", "download", "-x", "-modfile=tools/go.mod")
}

func InstallDeps() error {
	return Go("mod", "download", "-x")
}
