package main

import (
	"github.com/magefile/mage/sh"
)

type command string

var (
	ControllerGen = command("controller-gen")
	Gofumpt       = command("gofumpt")
	GolangciLint  = command("golangci-lint")
)

var tools map[command]string = map[command]string{
	ControllerGen: "sigs.k8s.io/controller-tools/cmd/controller-gen",
	Gofumpt:       "mvdan.cc/gofumpt",
	GolangciLint:  "github.com/golangci/golangci-lint/cmd/golangci-lint",
}

var (
	Go      = sh.RunCmd("go")
	Git     = sh.RunCmd("git")
	RunTool = RunVCmd("go", "run", "-modfile=tools/go.mod")
)

func RunVCmd(cmd string, primaryArgs ...string) func(args ...string) error {
	return func(args ...string) error {
		return sh.RunV(cmd, append(primaryArgs, args...)...)
	}
}
