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
	"os"
	"strings"

	"github.com/magefile/mage/sh"
)

func Test() error {
	out, err := OutTool(tools[Envtest], "use", k8sVersion, "--bin-dir", "bin", "-p", "path")
	if err != nil {
		return err
	}

	testEnv := map[string]string{
		"PATH": strings.Join([]string{os.Getenv("PATH"), out}, string(os.PathListSeparator)),
	}

	return sh.RunWithV(testEnv, "go", "run", "-modfile=tools/go.mod", tools[Gotestsum], "-f", "pkgname-and-test-fails", "--", "-race", "-shuffle=on", "./...")
}
