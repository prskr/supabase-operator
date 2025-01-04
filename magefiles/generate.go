package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"gopkg.in/yaml.v3"
)

const (
	composeFileUrl = "https://raw.githubusercontent.com/supabase/supabase/refs/heads/master/docker/docker-compose.yml"
)

func GenerateAll(ctx context.Context) {
	mg.CtxDeps(ctx, FetchImageMeta, FetchMigrations, CRDs, CRDDocs)
}

func CRDs() error {
	return errors.Join(
		RunTool(
			tools[ControllerGen],
			"rbac:roleName=manager-role",
			"crd",
			"webhook",
			`paths="./..."`,
			"output:crd:artifacts:config=config/crd/bases",
		),
		RunTool(tools[ControllerGen], `object:headerFile="hack/boilerplate.go.txt"`, `paths="./..."`),
	)
}

func CRDDocs() error {
	return RunTool(
		tools[CRDRefDocs],
		"--source-path=./api/",
		"--renderer=markdown",
		"--config=crd-docs.yaml",
		"--output-path=./docs/api/",
		"--output-mode=group",
	)
}

func FetchImageMeta(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, composeFileUrl, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var composeFile struct {
		Services map[string]struct {
			Image string `yaml:"image"`
		}
	}

	if err := yaml.NewDecoder(resp.Body).Decode(&composeFile); err != nil {
		return err
	}

	f, err := os.Create(filepath.Join("internal", "supabase", "images.go"))
	if err != nil {
		return err
	}

	defer f.Close()

	type imageRef struct {
		Repository string
		Tag        string
	}

	serviceMappings := map[string]string{
		"auth":      "Gotrue",
		"functions": "EdgeRuntime",
		"imgproxy":  "ImgProxy",
		"meta":      "PostgresMeta",
		"realtime":  "Realtime",
		"rest":      "Postgrest",
		"storage":   "Storage",
		"studio":    "Studio",
	}

	templateData := struct {
		Images map[string]imageRef
	}{
		Images: make(map[string]imageRef),
	}

	for name, service := range composeFile.Services {
		splitIdx := strings.LastIndex(service.Image, ":")
		repo := service.Image[:splitIdx]
		tag := service.Image[splitIdx+1:]

		mapping, ok := serviceMappings[name]
		if !ok {
			continue
		}

		templateData.Images[mapping] = imageRef{
			Repository: repo,
			Tag:        tag,
		}
	}

	latestEnvoyTag, err := latestReleaseVersion(ctx, "envoyproxy", "envoy")
	if err != nil {
		return err
	}

	templateData.Images["Envoy"] = imageRef{
		Repository: "envoyproxy/envoy",
		Tag:        fmt.Sprintf("distroless-%s", latestEnvoyTag),
	}

	if err := templates.ExecuteTemplate(f, "images.go.tmpl", templateData); err != nil {
		return err
	}

	if err := f.Sync(); err != nil {
		return err
	}

	return RunTool(tools[Gofumpt], "-l", "-w", f.Name())
}

func FetchMigrations(ctx context.Context) error {
	latestRelease, err := latestReleaseVersion(ctx, "supabase", "postgres")
	if err != nil {
		return err
	}

	checkoutDir, err := os.MkdirTemp(os.TempDir(), "supabase-*")
	if err != nil {
		return err
	}

	repoFS := os.DirFS(checkoutDir)

	defer os.RemoveAll(checkoutDir)

	if err := Git("clone", "--filter=blob:none", "--no-checkout", "https://github.com/supabase/postgres", checkoutDir); err != nil {
		return err
	}

	if err := Git("-C", checkoutDir, "sparse-checkout", "set", "--cone", "migrations"); err != nil {
		return err
	}

	if err := Git("-C", checkoutDir, "checkout", latestRelease); err != nil {
		return err
	}

	migrationsDirPath := path.Join(".", "migrations", "db")
	return fs.WalkDir(repoFS, migrationsDirPath, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || filepath.Ext(filePath) != ".sql" {
			return nil
		}

		fileName, err := filepath.Rel(migrationsDirPath, filePath)
		if err != nil {
			return err
		}

		dir, _ := filepath.Split(fileName)

		if err := os.MkdirAll(filepath.Join(workingDir, "assets", "migrations", dir), 0o750); err != nil {
			return err
		}

		slog.Info("Copying migration file", slog.String("file", fileName))
		return sh.Copy(filepath.Join(workingDir, "assets", "migrations", fileName), filepath.Join(checkoutDir, filepath.FromSlash(filePath)))
	})
}
