/*
Copyright 2024 Peter Kurfer.

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
	"context"
	"errors"
	"io"

	"github.com/jackc/pgx/v5"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
	"code.icb4dc0.de/prskr/supabase-operator/assets/migrations"
	"code.icb4dc0.de/prskr/supabase-operator/infrastructure/db"
)

// CoreReconciler reconciles a Core object
type CoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=cores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=cores/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=cores/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Core object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *CoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
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

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		logger.Error(err, "unable to connect to database")
		return ctrl.Result{}, err
	}

	defer CloseCtx(ctx, conn, &err)

	if err := r.applyMissingMigrations(ctx, conn, &core); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&supabasev1alpha1.Core{}).
		Owns(new(appsv1.Deployment)).
		Named("core").
		Complete(r)
}

func (r *CoreReconciler) applyMissingMigrations(ctx context.Context, conn *pgx.Conn, core *supabasev1alpha1.Core) (err error) {
	logger := log.FromContext(ctx)
	logger.Info("Checking for outstanding migrations")
	migrator := db.Migrator{Conn: conn}

	var appliedSomething bool

	if appliedSomething, err = migrator.ApplyAll(ctx, core.Status.AppliedMigrations, migrations.InitScripts()); err != nil {
		return err
	}

	if appliedSomething {
		logger.Info("Updating status after applying init scripts")
		return r.Client.Status().Update(ctx, core)
	}

	if appliedSomething, err = migrator.ApplyAll(ctx, core.Status.AppliedMigrations, migrations.MigrationScripts()); err != nil {
		return err
	}

	if appliedSomething {
		logger.Info("Updating status after applying migration scripts")
		return r.Client.Status().Update(ctx, core)
	}

	return nil
}

func Close(closer io.Closer, err *error) {
	*err = errors.Join(*err, closer.Close())
}

func CloseCtx(ctx context.Context, closable interface {
	Close(ctx context.Context) error
}, err *error,
) {
	*err = errors.Join(*err, closable.Close(ctx))
}
