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
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"iter"
	"path"
	"slices"
	"strings"
)

//go:embed setup/*.sql roles/*.sql init-scripts/*.sql migrations/*.sql
var MigrationsFS embed.FS

type Script struct {
	FileName string
	Content  string
	Hash     []byte
}

func InitScripts() iter.Seq2[Script, error] {
	return readScripts(path.Join(".", "init-scripts"))
}

func MigrationScripts() iter.Seq2[Script, error] {
	return readScripts(path.Join(".", "migrations"))
}

func RoleCreationScript(roleName string) (Script, error) {
	fileName := fmt.Sprintf("%s.sql", roleName)
	content, err := MigrationsFS.ReadFile(path.Join("roles", fileName))
	if err != nil {
		return Script{}, err
	}

	hash := sha256.New()
	_, _ = hash.Write(content)

	return Script{
		FileName: fileName,
		Content:  string(content),
		Hash:     hash.Sum(nil),
	}, nil
}

func readScripts(dir string) iter.Seq2[Script, error] {
	hash := sha256.New()
	return func(yield func(Script, error) bool) {
		files, err := MigrationsFS.ReadDir(dir)
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

			content, err := MigrationsFS.ReadFile(path.Join(dir, file.Name()))
			if err != nil {
				if !yield(Script{}, err) {
					return
				}
			}

			_, _ = hash.Write(content)

			s := Script{
				FileName: file.Name(),
				Content:  string(content),
				Hash:     hash.Sum(nil),
			}

			hash.Reset()

			if !yield(s, nil) {
				return
			}
		}
	}
}
