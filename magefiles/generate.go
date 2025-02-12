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
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"gopkg.in/yaml.v3"

	"code.icb4dc0.de/prskr/supabase-operator/internal/errx"
)

const (
	composeFileUrl = "https://raw.githubusercontent.com/supabase/supabase/refs/heads/master/docker/docker-compose.yml"
)

var ignoredMigrations = []string{
	"10000000000000_demote-postgres.sql",
}

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

func FetchImageMeta(ctx context.Context) (err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, composeFileUrl, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer errx.Close(resp.Body, &err)

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

	defer errx.Close(f, &err)

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
		Year   int
		Author string
	}{
		Images: make(map[string]imageRef),
		Year:   time.Now().Year(),
		Author: "Peter Kurfer",
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

func FetchMigrations(ctx context.Context) (err error) {
	latestRelease, err := latestReleaseVersion(ctx, "supabase", "postgres")
	if err != nil {
		return err
	}

	releaseArtifactURL := fmt.Sprintf("https://github.com/supabase/postgres/archive/refs/tags/%s.tar.gz", latestRelease)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releaseArtifactURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer errx.Close(resp.Body, &err)

	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}

	defer errx.Close(gzipReader, &err)

	migrationsDirPath := path.Join(fmt.Sprintf("postgres-%s", latestRelease), ".", "migrations", "db") + "/"
	tarReader := tar.NewReader(gzipReader)

	var header *tar.Header

	for header, err = tarReader.Next(); err == nil; header, err = tarReader.Next() {
		fileInfo := header.FileInfo()
		if fileInfo.IsDir() || path.Ext(fileInfo.Name()) != ".sql" {
			continue
		}

		fileName := header.Name
		if strings.HasPrefix(fileName, migrationsDirPath) {
			fileName = strings.TrimPrefix(fileName, migrationsDirPath)

			dir, migrationFileName := path.Split(fileName)

			if slices.Contains(ignoredMigrations, migrationFileName) {
				slog.Info("Skipping migration file", slog.String("name", migrationFileName))
				continue
			}

			outDir := filepath.Join(workingDir, "assets", "migrations", filepath.FromSlash(dir))
			if err := os.MkdirAll(outDir, 0o750); err != nil {
				return err
			}

			slog.Info("Copying file", slog.String("file", fileName))
			outFile, err := os.Create(filepath.Join(workingDir, "assets", "migrations", filepath.FromSlash(fileName)))
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}

			if err := outFile.Close(); err != nil {
				return err
			}

		} else {
			slog.Debug("skipping file", slog.String("file", fileName))
		}
	}

	if errors.Is(err, io.EOF) {
		return nil
	}

	return err
}
