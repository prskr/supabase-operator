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
