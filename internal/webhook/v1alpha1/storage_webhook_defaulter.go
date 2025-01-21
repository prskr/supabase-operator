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

package v1alpha1

import (
	"context"
	"fmt"
	"maps"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
	"code.icb4dc0.de/prskr/supabase-operator/internal/meta"
	"code.icb4dc0.de/prskr/supabase-operator/internal/pw"
)

// +kubebuilder:webhook:path=/mutate-supabase-k8s-icb4dc0-de-v1alpha1-storage,mutating=true,failurePolicy=fail,sideEffects=None,groups=supabase.k8s.icb4dc0.de,resources=storages,verbs=create;update,versions=v1alpha1,name=mstorage-v1alpha1.kb.io,admissionReviewVersions=v1

// StorageCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Storage when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type StorageCustomDefaulter struct {
	client.Client
}

var _ webhook.CustomDefaulter = &StorageCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Storage.
func (d *StorageCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	storage, ok := obj.(*supabasev1alpha1.Storage)

	if !ok {
		return fmt.Errorf("expected an Storage object but got %T", obj)
	}
	storagelog.Info("Defaulting for Storage", "name", storage.GetName())

	if err := d.defaultS3Protocol(ctx, storage); err != nil {
		return err
	}

	return nil
}

func (d *StorageCustomDefaulter) defaultS3Protocol(ctx context.Context, storage *supabasev1alpha1.Storage) error {
	if storage.Spec.S3 == nil {
		storage.Spec.S3 = new(supabasev1alpha1.S3ProtocolSpec)
	}

	if storage.Spec.S3.CredentialsSecretRef == nil {
		storage.Spec.S3.CredentialsSecretRef = &supabasev1alpha1.S3CredentialsRef{
			AccessKeyIdKey:     "accessKeyId",
			AccessSecretKeyKey: "secretAccessKey",
			SecretName:         fmt.Sprintf("%s-storage-protocol-s3-credentials", storage.Name),
		}
	}

	credentialsSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      storage.Spec.S3.CredentialsSecretRef.SecretName,
			Namespace: storage.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, d.Client, &credentialsSecret, func() error {
		credentialsSecret.Labels = maps.Clone(storage.Labels)
		if credentialsSecret.Labels == nil {
			credentialsSecret.Labels = make(map[string]string)
		}

		credentialsSecret.Labels[meta.SupabaseLabel.Reload] = ""

		if credentialsSecret.Data == nil {
			credentialsSecret.Data = make(map[string][]byte, 2)
		}

		if _, ok := credentialsSecret.Data[storage.Spec.S3.CredentialsSecretRef.AccessKeyIdKey]; !ok {
			credentialsSecret.Data[storage.Spec.S3.CredentialsSecretRef.AccessKeyIdKey] = pw.GeneratePW(32, nil)
		}

		if _, ok := credentialsSecret.Data[storage.Spec.S3.CredentialsSecretRef.AccessSecretKeyKey]; !ok {
			credentialsSecret.Data[storage.Spec.S3.CredentialsSecretRef.AccessSecretKeyKey] = pw.GeneratePW(64, nil)
		}

		return nil
	})

	return err
}
