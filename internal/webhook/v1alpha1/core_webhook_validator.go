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

package v1alpha1

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
	"code.icb4dc0.de/prskr/supabase-operator/internal/supabase"
)

// nolint:unused
// log is for logging in this package.
var corelog = logf.Log.WithName("core-resource")

var (
	ErrNoDSN                           = errors.New("neither DSN nor DSNFrom are set - either one needs to be specified")
	ErrManagedCredentialsNotSpecified  = errors.New("credentials are not set which is required when self managing DB roles")
	ErrManagedCredentialsSecretMissing = errors.New("secret does not exist")
)

// +kubebuilder:webhook:path=/validate-supabase-k8s-icb4dc0-de-v1alpha1-core,mutating=false,failurePolicy=fail,sideEffects=None,groups=supabase.k8s.icb4dc0.de,resources=cores,verbs=create;update,versions=v1alpha1,name=vcore-v1alpha1.kb.io,admissionReviewVersions=v1

// CoreCustomValidator struct is responsible for validating the Core resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type CoreCustomValidator struct {
	client.Client
}

var _ webhook.CustomValidator = &CoreCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Core.
func (v *CoreCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	core, ok := obj.(*supabasev1alpha1.Core)
	if !ok {
		return nil, fmt.Errorf("expected a Core object but got %T", obj)
	}
	corelog.Info("Validation for Core upon creation", "name", core.GetName())

	warns, err := v.validateDb(ctx, core)
	if err != nil {
		return nil, err
	}

	return warns, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Core.
func (v *CoreCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	core, ok := newObj.(*supabasev1alpha1.Core)
	if !ok {
		return nil, fmt.Errorf("expected a Core object for the newObj but got %T", newObj)
	}
	corelog.Info("Validation for Core upon update", "name", core.GetName())

	warns, err := v.validateDb(ctx, core)
	if err != nil {
		return nil, err
	}

	return warns, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Core.
func (v *CoreCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	core, ok := obj.(*supabasev1alpha1.Core)
	if !ok {
		return nil, fmt.Errorf("expected a Core object but got %T", obj)
	}
	corelog.Info("Validation for Core upon deletion", "name", core.GetName())

	warns, err := v.validateDb(ctx, core)
	if err != nil {
		return nil, err
	}

	return warns, nil
}

func (v *CoreCustomValidator) validateDb(
	ctx context.Context,
	core *supabasev1alpha1.Core,
) (warnings admission.Warnings, err error) {
	dbSpec := core.Spec.Database

	if dbSpec.DSN != nil && dbSpec.DSNSecretRef == nil {
		return warnings, ErrNoDSN
	}

	if dbSpec.Roles.SelfManaged {
		doesSecretExists := func(ctx context.Context, name string) (exists bool, err error) {
			var secret corev1.Secret
			if err := v.Client.Get(ctx, types.NamespacedName{Namespace: core.Namespace, Name: name}, &secret); err == nil {
				return true, nil
			} else if client.IgnoreNotFound(err) == nil {
				return false, nil
			} else {
				return false, err
			}
		}

		if authenticator := dbSpec.Roles.Secrets.Authenticator; authenticator == nil {
			return warnings, fmt.Errorf("%w: %s", ErrManagedCredentialsNotSpecified, supabase.DBRoleAuthenticator)
		} else {
			exists, err := doesSecretExists(ctx, authenticator.Name)
			if err != nil {
				return warnings, err
			} else if !exists {
				return warnings, fmt.Errorf("%w: %s", ErrManagedCredentialsSecretMissing, authenticator.Name)
			}
		}

		if authAdmin := dbSpec.Roles.Secrets.AuthAdmin; authAdmin == nil {
			return warnings, fmt.Errorf("%w: %s", ErrManagedCredentialsNotSpecified, supabase.DBRoleAuthAdmin)
		} else {
			exists, err := doesSecretExists(ctx, authAdmin.Name)
			if err != nil {
				return warnings, err
			} else if !exists {
				return warnings, fmt.Errorf("%w: %s", ErrManagedCredentialsSecretMissing, authAdmin.Name)
			}
		}

		if functionsAdmin := dbSpec.Roles.Secrets.FunctionsAdmin; functionsAdmin == nil {
			return warnings, fmt.Errorf("%w: %s", ErrManagedCredentialsNotSpecified, supabase.DBRoleFunctionsAdmin)
		} else {
			exists, err := doesSecretExists(ctx, functionsAdmin.Name)
			if err != nil {
				return warnings, err
			} else if !exists {
				return warnings, fmt.Errorf("%w: %s", ErrManagedCredentialsSecretMissing, functionsAdmin.Name)
			}
		}

		if storageAdmin := dbSpec.Roles.Secrets.StorageAdmin; storageAdmin == nil {
			return warnings, fmt.Errorf("%w: %s", ErrManagedCredentialsNotSpecified, supabase.DBRoleStorageAdmin)
		} else {
			exists, err := doesSecretExists(ctx, storageAdmin.Name)
			if err != nil {
				return warnings, err
			} else if !exists {
				return warnings, fmt.Errorf("%w: %s", ErrManagedCredentialsSecretMissing, storageAdmin.Name)
			}
		}
	}

	return warnings, nil
}
