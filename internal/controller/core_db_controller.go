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

package controller

import (
	"bytes"
	"context"
	"hash/fnv"
	"maps"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
	"code.icb4dc0.de/prskr/supabase-operator/assets/migrations"
	"code.icb4dc0.de/prskr/supabase-operator/internal/db"
	"code.icb4dc0.de/prskr/supabase-operator/internal/errx"
	"code.icb4dc0.de/prskr/supabase-operator/internal/meta"
	"code.icb4dc0.de/prskr/supabase-operator/internal/pw"
	"code.icb4dc0.de/prskr/supabase-operator/internal/supabase"
)

// CoreDbReconciler reconciles a Core object
type CoreDbReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *CoreDbReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	logger := log.FromContext(ctx)

	var core supabasev1alpha1.Core

	if err := r.Get(ctx, req.NamespacedName, &core); client.IgnoreNotFound(err) != nil {
		logger.Error(err, "unable to fetch Core")
		return ctrl.Result{}, err
	}

	dsn, err := core.Spec.Database.GetDSN(ctx, client.NewNamespacedClient(r.Client, req.Namespace))
	if err != nil {
		logger.Error(err, "unable to get DSN")
		return ctrl.Result{}, err
	}

	logger.Info("Connecting to database")
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		logger.Error(err, "unable to connect to database")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	defer errx.CloseCtx(ctx, conn, &err)

	logger.Info("Connected to database, checking for outstanding migrations")
	if err := r.applyMissingMigrations(ctx, conn, &core); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Sync credentials for Supabase roles")
	if err := r.ensureDbRolesSecrets(ctx, dsn, conn, &core); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CoreDbReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(new(supabasev1alpha1.Core)).
		Owns(new(corev1.Secret)).
		Named("core-db").
		Complete(r)
}

func (r *CoreDbReconciler) applyMissingMigrations(
	ctx context.Context,
	conn *pgx.Conn,
	core *supabasev1alpha1.Core,
) (err error) {
	logger := log.FromContext(ctx)
	migrator := db.Migrator{Conn: conn}

	var appliedSomething bool

	if core.Status.Database.AppliedMigrations == nil {
		core.Status.Database.AppliedMigrations = make(supabasev1alpha1.MigrationStatus)
	}

	if appliedSomething, err = migrator.ApplyAll(ctx, core.Status.Database.AppliedMigrations, migrations.InitScripts()); err != nil {
		return err
	}

	if appliedSomething {
		logger.Info("Updating status after applying init scripts")
		return r.Client.Status().Update(ctx, core)
	} else {
		logger.Info("Init scripts were up to date - did not run any")
	}

	if appliedSomething, err = migrator.ApplyAll(ctx, core.Status.Database.AppliedMigrations, migrations.MigrationScripts()); err != nil {
		return err
	}

	if appliedSomething {
		logger.Info("Updating status after applying migration scripts")
		return r.Client.Status().Update(ctx, core)
	} else {
		logger.Info("Migrrations were up to date - did not run any")
	}

	return nil
}

func (r *CoreDbReconciler) ensureDbRolesSecrets(
	ctx context.Context,
	dsn string,
	conn *pgx.Conn,
	core *supabasev1alpha1.Core,
) error {
	var (
		logger   = log.FromContext(ctx)
		rolesMgr = db.NewRolesManager(conn)
	)

	dbSpec := core.Spec.Database
	if dbSpec.Roles.SelfManaged {
		logger.Info("Database roles are self-managed, skipping reconciliation")
		return nil
	}

	parsedDSN, err := url.Parse(dsn)
	if err != nil {
		return err
	}

	var (
		dsnUser  = parsedDSN.User.Username()
		dsnPW, _ = parsedDSN.User.Password()
	)

	roles := map[string]supabase.DBRole{
		dbSpec.Roles.Secrets.Authenticator:  supabase.DBRoleAuthenticator,
		dbSpec.Roles.Secrets.AuthAdmin:      supabase.DBRoleAuthAdmin,
		dbSpec.Roles.Secrets.FunctionsAdmin: supabase.DBRoleFunctionsAdmin,
		dbSpec.Roles.Secrets.StorageAdmin:   supabase.DBRoleStorageAdmin,
		dbSpec.Roles.Secrets.Admin:          supabase.DBRoleSupabaseAdmin,
	}

	if core.Status.Database.Roles == nil {
		core.Status.Database.Roles = make(map[string][]byte)
	}

	hash := fnv.New64a()

	for secretName, role := range roles {
		secretLogger := logger.WithValues("secret_name", secretName, "role_name", role.String())

		secretLogger.Info("Ensuring credential secret")

		credentialsSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: core.Namespace,
			},
		}

		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, credentialsSecret, func() error {
			logger.Info("Ensuring role credentials", "role_name", role.String())

			credentialsSecret.Labels = maps.Clone(core.Labels)
			if credentialsSecret.Labels == nil {
				credentialsSecret.Labels = make(map[string]string)
			}

			credentialsSecret.Labels[meta.SupabaseLabel.Reload] = ""

			if credentialsSecret.Data == nil {
				credentialsSecret.Data = make(map[string][]byte)
			}

			if _, ok := credentialsSecret.Data[corev1.BasicAuthUsernameKey]; !ok {
				credentialsSecret.Data[corev1.BasicAuthUsernameKey] = role.Bytes()
			}

			var requireStatusUpdate bool

			if value := credentialsSecret.Data[corev1.BasicAuthPasswordKey]; len(value) == 0 || (role.String() == dsnUser && !bytes.Equal(credentialsSecret.Data[corev1.BasicAuthPasswordKey], []byte(dsnPW))) {
				if role.String() == dsnUser {
					credentialsSecret.Data[corev1.BasicAuthPasswordKey] = []byte(dsnPW)
				} else {
					credentialsSecret.Data[corev1.BasicAuthPasswordKey] = pw.GeneratePW(24, nil)
				}

				secretLogger.Info("Update database role to match secret credentials")
				if err := rolesMgr.UpdateRolePassword(ctx, role.String(), credentialsSecret.Data[corev1.BasicAuthPasswordKey]); err != nil {
					return err
				}
				core.Status.Database.Roles[role.String()] = hash.Sum(credentialsSecret.Data[corev1.BasicAuthPasswordKey])
				requireStatusUpdate = true
			} else {
				if bytes.Equal(core.Status.Database.Roles[role.String()], hash.Sum(credentialsSecret.Data[corev1.BasicAuthPasswordKey])) {
					logger.Info("Role password is up to date", "role_name", role.String())
				} else {
					if err := rolesMgr.UpdateRolePassword(ctx, role.String(), credentialsSecret.Data[corev1.BasicAuthPasswordKey]); err != nil {
						return err
					}
					requireStatusUpdate = true
				}
				core.Status.Database.Roles[role.String()] = hash.Sum(credentialsSecret.Data[corev1.BasicAuthPasswordKey])
			}

			credentialsSecret.Type = corev1.SecretTypeBasicAuth

			if requireStatusUpdate {
				secretLogger.Info("Updating status")
				if err := r.Status().Update(ctx, core); err != nil {
					return err
				}
			}

			logger.Info("Setting owner reference for credentials secret")
			if err := controllerutil.SetControllerReference(core, credentialsSecret, r.Scheme); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}
