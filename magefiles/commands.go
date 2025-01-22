/*
Copyright 2025 Peter Kurfer.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/magefile/mage/sh"
)

type command string

var (
	ControllerGen = command("controller-gen")
	Gofumpt       = command("gofumpt")
	GolangciLint  = command("golangci-lint")
	Gotestsum     = command("gotestsum")
	Sqlc          = command("sqlc")
	CRDRefDocs    = command("crd-ref-docs")
	Envtest       = command("envtest")
)

var tools map[command]string = map[command]string{
	ControllerGen: "sigs.k8s.io/controller-tools/cmd/controller-gen",
	Gofumpt:       "mvdan.cc/gofumpt",
	GolangciLint:  "github.com/golangci/golangci-lint/cmd/golangci-lint",
	Gotestsum:     "gotest.tools/gotestsum",
	Sqlc:          "github.com/sqlc-dev/sqlc/cmd/sqlc",
	CRDRefDocs:    "github.com/elastic/crd-ref-docs",
	Envtest:       "sigs.k8s.io/controller-runtime/tools/setup-envtest",
}

var (
	Go      = sh.RunCmd("go")
	Git     = sh.RunCmd("git")
	RunTool = RunVCmd("go", "run", "-modfile=tools/go.mod")
	OutTool = sh.OutCmd("go", "run", "-modfile=tools/go.mod")
)

func RunVCmd(cmd string, primaryArgs ...string) func(args ...string) error {
	return func(args ...string) error {
		return sh.RunV(cmd, append(primaryArgs, args...)...)
	}
}
