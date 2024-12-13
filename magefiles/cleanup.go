package main

import "github.com/magefile/mage/mg"

func Validate() {
	mg.Deps(Fmt, Lint)
}

func Fmt() error {
	return RunTool(tools[Gofumpt], "-l", "-w", ".")
}

func Vet() error {
	return Go("vet", "./...")
}

func Lint() error {
	return RunTool(tools[GolangciLint], "run")
}

func LintFix() error {
	return RunTool(tools[GolangciLint], "run", "--fix")
}
