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

package migrations

import (
	"embed"
	"fmt"
	"io/fs"
	"iter"
	"path"
	"slices"
	"strings"
)

//go:embed */*.sql
var migrationsFS embed.FS

type Script struct {
	FileName string
	Content  string
}

func InitScripts() iter.Seq2[Script, error] {
	return readScripts(path.Join(".", "init-scripts"))
}

func MigrationScripts() iter.Seq2[Script, error] {
	return readScripts(path.Join(".", "migrations"))
}

func RoleCreationScript(roleName string) (Script, error) {
	fileName := fmt.Sprintf("%s.sql", roleName)
	content, err := migrationsFS.ReadFile(path.Join("roles", fileName))
	if err != nil {
		return Script{}, err
	}

	return Script{fileName, string(content)}, nil
}

func readScripts(dir string) iter.Seq2[Script, error] {
	return func(yield func(Script, error) bool) {
		files, err := migrationsFS.ReadDir(dir)
		if err != nil {
			yield(Script{}, err)
			return
		}

		slices.SortFunc(files, func(a, b fs.DirEntry) int {
			return strings.Compare(a.Name(), b.Name())
		})

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			content, err := migrationsFS.ReadFile(path.Join(dir, file.Name()))
			if err != nil {
				if !yield(Script{}, err) {
					return
				}
			}

			s := Script{
				FileName: file.Name(),
				Content:  string(content),
			}

			if !yield(s, nil) {
				return
			}
		}
	}
}
