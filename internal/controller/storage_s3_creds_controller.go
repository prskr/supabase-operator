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
	"context"
	"errors"
	"maps"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	supabasev1alpha1 "github.com/prskr/supabase-operator/api/v1alpha1"
	"github.com/prskr/supabase-operator/internal/meta"
	"github.com/prskr/supabase-operator/internal/pw"
)

var errS3CredentialsSecretEmpty = errors.New("S3 storage credentials secret is empty")

type StorageS3CredentialsReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *StorageS3CredentialsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	var (
		storage supabasev1alpha1.Storage
		logger  = log.FromContext(ctx)
	)

	if err := r.Get(ctx, req.NamespacedName, &storage); err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Info("Storage instance does not exist")
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if err := r.reconcileS3ProtoSecret(ctx, &storage); err != nil {
		return ctrl.Result{}, err
	}

	if storage.Spec.API.S3Backend != nil {
		if err := r.reconcileS3StorageSecret(ctx, &storage); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StorageS3CredentialsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(new(supabasev1alpha1.Storage)).
		Owns(new(corev1.Secret)).
		Named("storage-s3-creds").
		Complete(r)
}

func (r *StorageS3CredentialsReconciler) reconcileS3StorageSecret(
	ctx context.Context,
	storage *supabasev1alpha1.Storage,
) error {
	if storage.Spec.API.S3Backend.CredentialsSecretRef == nil {
		return errS3CredentialsSecretEmpty
	}

	s3CredsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      storage.Spec.API.S3Backend.CredentialsSecretRef.SecretName,
			Namespace: storage.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, s3CredsSecret, func() error {
		return controllerutil.SetControllerReference(storage, s3CredsSecret, r.Scheme)
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *StorageS3CredentialsReconciler) reconcileS3ProtoSecret(
	ctx context.Context,
	storage *supabasev1alpha1.Storage,
) error {
	const (
		acccessKeyIdAndSecret = 2
	)
	s3ProtoSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      storage.Spec.API.S3Protocol.CredentialsSecretRef.SecretName,
			Namespace: storage.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &s3ProtoSecret, func() error {
		s3ProtoSecret.Labels = maps.Clone(storage.Labels)
		if s3ProtoSecret.Labels == nil {
			s3ProtoSecret.Labels = make(map[string]string)
		}

		s3ProtoSecret.Labels[meta.SupabaseLabel.Reload] = ""

		if err := controllerutil.SetControllerReference(storage, &s3ProtoSecret, r.Scheme); err != nil {
			return err
		}

		if s3ProtoSecret.Data == nil {
			s3ProtoSecret.Data = make(map[string][]byte, acccessKeyIdAndSecret)
		}

		if _, ok := s3ProtoSecret.Data[storage.Spec.API.S3Protocol.CredentialsSecretRef.AccessKeyIDKey]; !ok {
			s3ProtoSecret.Data[storage.Spec.API.S3Protocol.CredentialsSecretRef.AccessKeyIDKey] = pw.GeneratePW(32, nil)
		}

		if _, ok := s3ProtoSecret.Data[storage.Spec.API.S3Protocol.CredentialsSecretRef.AccessSecretKeyKey]; !ok {
			s3ProtoSecret.Data[storage.Spec.API.S3Protocol.CredentialsSecretRef.AccessSecretKeyKey] = pw.GeneratePW(64, nil)
		}

		return nil
	})

	return err
}
