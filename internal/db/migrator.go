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

package db

import (
	"context"
	"errors"
	"iter"

	"github.com/jackc/pgx/v5"
	"sigs.k8s.io/controller-runtime/pkg/log"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
	"code.icb4dc0.de/prskr/supabase-operator/assets/migrations"
)

type Migrator struct {
	Conn *pgx.Conn
}

func (m Migrator) ApplyAll(
	ctx context.Context,
	status *supabasev1alpha1.CoreStatus,
	seq iter.Seq2[migrations.Script, error],
	areInitScripts bool,
) (appliedSomething bool, err error) {
	logger := log.FromContext(ctx)

	for s, err := range seq {
		if err != nil {
			return false, err
		}

		if found, upToDate := status.Database.IsMigrationUpToDate(s.FileName, s.Hash); found && upToDate {
			continue
		} else if found && !upToDate && areInitScripts {
			logger.Info("Change in init script was detected - will not apply because init scripts are not idempotent", "file_name", s.FileName)
			continue
		}

		logger.Info("Applying missing or outdated migration", "filename", s.FileName)
		err := status.Database.RecordMigrationCondition(s.FileName, s.Hash, m.Apply(ctx, s.Content))
		if err != nil {
			return false, err
		}

		appliedSomething = true
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
