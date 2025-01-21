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

	"k8s.io/apimachinery/pkg/runtime"
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
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &StorageCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Storage.
func (v *StorageCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	storage, ok := obj.(*supabasev1alpha1.Storage)
	if !ok {
		return nil, fmt.Errorf("expected a Storage object but got %T", obj)
	}
	storagelog.Info("Validation for Storage upon creation", "name", storage.GetName())

	// TODO(user): fill in your validation logic upon object creation.

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Storage.
func (v *StorageCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	storage, ok := newObj.(*supabasev1alpha1.Storage)
	if !ok {
		return nil, fmt.Errorf("expected a Storage object for the newObj but got %T", newObj)
	}
	storagelog.Info("Validation for Storage upon update", "name", storage.GetName())

	// TODO(user): fill in your validation logic upon object update.

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Storage.
func (v *StorageCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	storage, ok := obj.(*supabasev1alpha1.Storage)
	if !ok {
		return nil, fmt.Errorf("expected a Storage object but got %T", obj)
	}
	storagelog.Info("Validation for Storage upon deletion", "name", storage.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
