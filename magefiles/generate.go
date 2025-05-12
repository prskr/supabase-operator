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

	"github.com/magefile/mage/mg"

	"code.icb4dc0.de/prskr/supabase-operator/internal/errx"
)

const (
	composeFileUrl = "https://raw.githubusercontent.com/supabase/supabase/refs/heads/master/docker/docker-compose.yml"
)

var ignoredMigrations = []string{
	"10000000000000_demote-postgres.sql",
	"20250312095419_pgbouncer_ownership.sql",
}

func GenerateAll(ctx context.Context) {
	mg.CtxDeps(ctx, FetchMigrations)
}

func FetchMigrations(ctx context.Context) (err error) {
	latestRelease, err := latestReleaseVersion(ctx, "supabase", "postgres", excludeDrafts, excludePreReleases, matchesTagPattern(`15\..*`))
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "Extracting Postgres migrations for release", slog.String("release", latestRelease))

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

		if header == nil {
			slog.Warn("header is nil - what's happening?!")
			continue
		}

		fileName := header.Name
		if after, ok := strings.CutPrefix(fileName, migrationsDirPath); ok {
			fileName = after

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
