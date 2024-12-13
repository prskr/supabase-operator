package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"

	"code.icb4dc0.de/prskr/supabase-operator/assets/migrations"
)

func Migrate(ctx context.Context) error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return errors.New("DATABASE_URL is required")
	}

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return err
	}

	defer conn.Close(ctx)

	for s, err := range migrations.InitScripts() {
		if err != nil {
			return err
		}

		slog.Info("Running init script", slog.String("file", s.FileName))

		_, err = conn.Exec(ctx, s.Content)
		if err != nil {
			return err
		}
	}

	for s, err := range migrations.MigrationScripts() {
		if err != nil {
			return err
		}

		slog.Info("Running migration script", slog.String("file", s.FileName))

		_, err = conn.Exec(ctx, s.Content)
		if err != nil {
			return err
		}
	}

	return nil
}
