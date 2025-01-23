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
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var storagelog = logf.Log.WithName("storage-resource")

// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-supabase-k8s-icb4dc0-de-v1alpha1-storage,mutating=false,failurePolicy=fail,sideEffects=None,groups=supabase.k8s.icb4dc0.de,resources=storages,verbs=create;update,versions=v1alpha1,name=vstorage-v1alpha1.kb.io,admissionReviewVersions=v1

// StorageCustomValidator struct is responsible for validating the Storage resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type StorageCustomValidator struct {
	client.Client
}

var _ webhook.CustomValidator = &StorageCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Storage.
func (v *StorageCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	storage, ok := obj.(*supabasev1alpha1.Storage)
	if !ok {
		return nil, fmt.Errorf("expected a Storage object but got %T", obj)
	}
	storagelog.Info("Validation for Storage upon creation", "name", storage.GetName())

	if ws, err := v.validateStorageApi(ctx, storage); err != nil {
		return ws, err
	} else {
		warnings = append(warnings, ws...)
	}

	return warnings, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Storage.
func (v *StorageCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (warnings admission.Warnings, err error) {
	storage, ok := newObj.(*supabasev1alpha1.Storage)
	if !ok {
		return nil, fmt.Errorf("expected a Storage object for the newObj but got %T", newObj)
	}
	storagelog.Info("Validation for Storage upon update", "name", storage.GetName())

	if ws, err := v.validateStorageApi(ctx, storage); err != nil {
		return ws, err
	} else {
		warnings = append(warnings, ws...)
	}

	return warnings, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Storage.
func (v *StorageCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	storage, ok := obj.(*supabasev1alpha1.Storage)
	if !ok {
		return nil, fmt.Errorf("expected a Storage object but got %T", obj)
	}
	storagelog.Info("Validation for Storage upon deletion", "name", storage.GetName())

	return nil, nil
}

func (v *StorageCustomValidator) validateStorageApi(ctx context.Context, storage *supabasev1alpha1.Storage) (admission.Warnings, error) {
	var warnings admission.Warnings

	apiSpec := storage.Spec.Api

	if (apiSpec.FileBackend == nil) == (apiSpec.S3Backend == nil) {
		return nil, errors.New("it is not possible to configure both or non backend at all - please configure either file or S3 backend")
	}

	if apiSpec.S3Backend != nil {
		if apiSpec.S3Backend.CredentialsSecretRef == nil {
			return nil, errors.New(".spec.api.s3Backend.credentialsSecretRef cannot be empty")
		}

		s3CredentialsSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: apiSpec.S3Backend.CredentialsSecretRef.SecretName,
			},
		}

		if err := v.Get(ctx, client.ObjectKeyFromObject(s3CredentialsSecret), s3CredentialsSecret); err != nil {
			if client.IgnoreNotFound(err) == nil {
				warnings = append(warnings, fmt.Sprintf("Secret %q could not be found", apiSpec.S3Backend.CredentialsSecretRef.SecretName))
			} else {
				return nil, err
			}
		} else {
			if accessKeyId, ok := s3CredentialsSecret.Data[apiSpec.S3Backend.CredentialsSecretRef.AccessKeyIdKey]; !ok {
				return warnings, fmt.Errorf("secret %q does not contain an access key id at specified key %q", apiSpec.S3Backend.CredentialsSecretRef.SecretName, apiSpec.S3Backend.CredentialsSecretRef.AccessKeyIdKey)
			} else if len(accessKeyId) == 0 {
				return warnings, fmt.Errorf("access key id in Secret %q with key %q is empty", apiSpec.S3Backend.CredentialsSecretRef.SecretName, apiSpec.S3Backend.CredentialsSecretRef.AccessKeyIdKey)
			}

			if accessSecretKey, ok := s3CredentialsSecret.Data[apiSpec.S3Backend.CredentialsSecretRef.AccessSecretKeyKey]; !ok {
				return warnings, fmt.Errorf("secret %q does not contain an access secret key at specified key %q", apiSpec.S3Backend.CredentialsSecretRef.SecretName, apiSpec.S3Backend.CredentialsSecretRef.AccessSecretKeyKey)
			} else if len(accessSecretKey) == 0 {
				return warnings, fmt.Errorf("access secret key in Secret %q with key %q is empty", apiSpec.S3Backend.CredentialsSecretRef.SecretName, apiSpec.S3Backend.CredentialsSecretRef.AccessSecretKeyKey)
			}
		}
	}

	return warnings, nil
}
