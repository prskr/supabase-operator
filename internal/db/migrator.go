package db

import (
	"context"
	"errors"
	"iter"

	"code.icb4dc0.de/prskr/supabase-operator/assets/migrations"
	"github.com/jackc/pgx/v5"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
)

type Migrator struct {
	Conn *pgx.Conn
}

func (m Migrator) ApplyAll(ctx context.Context, status supabasev1alpha1.MigrationStatus, seq iter.Seq2[migrations.Script, error]) (appliedSomething bool, err error) {
	for s, err := range seq {
		if err != nil {
			return false, err
		}

		if status.IsApplied(s.FileName) {
			continue
		}

		if err := m.Apply(ctx, s.Content); err != nil {
			return false, err
		}

		appliedSomething = true
		status.Record(s.FileName)
	}

	return appliedSomething, nil
}

func (m Migrator) Apply(ctx context.Context, script string) error {
	tx, err := m.Conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, script)
	if err != nil {
		return errors.Join(err, tx.Rollback(ctx))
	}

	return tx.Commit(ctx)
}
