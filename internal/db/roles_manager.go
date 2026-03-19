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
	"fmt"

	"github.com/jackc/pgx/v5"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/prskr/supabase-operator/assets/migrations"
)

const (
	alterUserPwd    = `ALTER ROLE %s WITH PASSWORD '%s';`
	checkUserExists = `SELECT 1 FROM pg_user WHERE usename = $1;`
)

func NewRolesManager(conn *pgx.Conn) RolesManager {
	return RolesManager{
		Conn: conn,
	}
}

type RolesManager struct {
	Conn *pgx.Conn
}

func (mgr RolesManager) UpdateRolePassword(ctx context.Context, roleName string, password []byte) error {
	if err := mgr.ensureLoginRoleExists(ctx, roleName); err != nil {
		return err
	}

	_, err := mgr.Conn.Exec(ctx, fmt.Sprintf(alterUserPwd, roleName, password))
	return err
}

func (mgr RolesManager) ensureLoginRoleExists(ctx context.Context, roleName string) error {
	logger := log.FromContext(ctx).WithValues("role_name", roleName)

	rows, err := mgr.Conn.Query(ctx, checkUserExists, roleName)
	if err != nil {
		return err
	}

	defer rows.Close()

	_, err = pgx.CollectExactlyOneRow(rows, func(row pgx.CollectableRow) (out int, err error) {
		err = row.Scan(&out)
		return
	})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		logger.Info("No rows, this means the role does not exists, creating it now")
	} else {
		return nil
	}

	script, err := migrations.RoleCreationScript(roleName)
	if err != nil {
		return err
	}

	_, err = mgr.Conn.Exec(ctx, script.Content)

	return err
}
