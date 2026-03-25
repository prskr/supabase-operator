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

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	supabasev1alpha1 "github.com/prskr/supabase-operator/api/v1alpha1"
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

var _ admission.Defaulter[*supabasev1alpha1.Storage] = &StorageCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Storage.
func (d *StorageCustomDefaulter) Default(ctx context.Context, storage *supabasev1alpha1.Storage) error {
	storagelog.Info("Defaulting for Storage", "name", storage.GetName())

	d.defaultS3Protocol(storage)

	return nil
}

func (d *StorageCustomDefaulter) defaultS3Protocol(storage *supabasev1alpha1.Storage) {
	if storage.Spec.API.S3Protocol == nil {
		storage.Spec.API.S3Protocol = new(supabasev1alpha1.S3ProtocolSpec)
	}

	if storage.Spec.API.S3Protocol.CredentialsSecretRef == nil {
		storage.Spec.API.S3Protocol.CredentialsSecretRef = &supabasev1alpha1.S3CredentialsRef{
			SecretName: fmt.Sprintf("%s-storage-protocol-s3-credentials", storage.Name),
		}
	}
}
