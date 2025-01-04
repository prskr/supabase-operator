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
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
	"code.icb4dc0.de/prskr/supabase-operator/internal/jwk"
	"code.icb4dc0.de/prskr/supabase-operator/internal/meta"
	"code.icb4dc0.de/prskr/supabase-operator/internal/supabase"
)

// +kubebuilder:webhook:path=/mutate-supabase-k8s-icb4dc0-de-v1alpha1-core,mutating=true,failurePolicy=fail,sideEffects=None,groups=supabase.k8s.icb4dc0.de,resources=cores,verbs=create;update,versions=v1alpha1,name=mcore-v1alpha1.kb.io,admissionReviewVersions=v1

// CoreCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Core when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type CoreCustomDefaulter struct {
	client.Client
}

var _ webhook.CustomDefaulter = &CoreCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Core.
func (d *CoreCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	core, ok := obj.(*supabasev1alpha1.Core)

	if !ok {
		return fmt.Errorf("expected an Core object but got %T", obj)
	}
	corelog.Info("Defaulting for Core", "name", core.GetName())

	if err := d.defaultJWT(ctx, core); err != nil {
		return fmt.Errorf("ensuring JWT secret: %w", err)
	}

	// TODO copy plain text DSN to secret if present

	corelog.Info("Defaulting database roles")
	if !core.Spec.Database.Roles.SelfManaged {
		const roleCredsSecretNameTemplate = "db-roles-creds-%s"
		if core.Spec.Database.Roles.Secrets.Admin == nil {
			corelog.Info("Defaulting role", "role_name", supabase.DBRoleSupabaseAdmin)
			core.Spec.Database.Roles.Secrets.Admin = &corev1.LocalObjectReference{
				Name: fmt.Sprintf(roleCredsSecretNameTemplate, supabase.DBRoleSupabaseAdmin.K8sString()),
			}
		}

		if core.Spec.Database.Roles.Secrets.Authenticator == nil {
			corelog.Info("Defaulting role", "role_name", supabase.DBRoleAuthenticator)
			core.Spec.Database.Roles.Secrets.Authenticator = &corev1.LocalObjectReference{
				Name: fmt.Sprintf(roleCredsSecretNameTemplate, supabase.DBRoleAuthenticator.K8sString()),
			}
		}

		if core.Spec.Database.Roles.Secrets.AuthAdmin == nil {
			corelog.Info("Defaulting role", "role_name", supabase.DBRoleAuthAdmin)
			core.Spec.Database.Roles.Secrets.AuthAdmin = &corev1.LocalObjectReference{
				Name: fmt.Sprintf(roleCredsSecretNameTemplate, supabase.DBRoleAuthAdmin.K8sString()),
			}
		}

		if core.Spec.Database.Roles.Secrets.FunctionsAdmin == nil {
			corelog.Info("Defaulting role", "role_name", supabase.DBRoleFunctionsAdmin)
			core.Spec.Database.Roles.Secrets.FunctionsAdmin = &corev1.LocalObjectReference{
				Name: fmt.Sprintf(roleCredsSecretNameTemplate, supabase.DBRoleFunctionsAdmin.K8sString()),
			}
		}

		if core.Spec.Database.Roles.Secrets.StorageAdmin == nil {
			corelog.Info("Defaulting role", "role_name", supabase.DBRoleStorageAdmin)
			core.Spec.Database.Roles.Secrets.StorageAdmin = &corev1.LocalObjectReference{
				Name: fmt.Sprintf(roleCredsSecretNameTemplate, supabase.DBRoleStorageAdmin.K8sString()),
			}
		}
	}

	return nil
}

func (d *CoreCustomDefaulter) defaultJWT(ctx context.Context, core *supabasev1alpha1.Core) error {
	corelog.Info("Defaulting JWT")

	if core.Spec.JWT == nil {
		core.Spec.JWT = new(supabasev1alpha1.JwtSpec)
	}

	if core.Spec.JWT.SecretRef == nil {
		core.Spec.JWT.SecretRef = &corev1.LocalObjectReference{
			Name: supabase.ServiceConfig.JWT.ObjectName(core),
		}
	}

	if core.Spec.JWT.SecretKey == "" {
		core.Spec.JWT.SecretKey = supabase.ServiceConfig.JWT.Defaults.SecretKey
	}

	if core.Spec.JWT.JwksKey == "" {
		core.Spec.JWT.JwksKey = supabase.ServiceConfig.JWT.Defaults.JwksKey
	}

	if core.Spec.JWT.AnonKey == "" {
		core.Spec.JWT.AnonKey = supabase.ServiceConfig.JWT.Defaults.AnonKey
	}

	if core.Spec.JWT.ServiceKey == "" {
		core.Spec.JWT.ServiceKey = supabase.ServiceConfig.JWT.Defaults.ServiceKey
	}

	jwtSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      core.Spec.JWT.SecretRef.Name,
			Namespace: core.Namespace,
		},
	}

	if core.Spec.JWT.Secret == nil {
		return nil
	}

	_, err := controllerutil.CreateOrUpdate(ctx, d.Client, jwtSecret, func() error {
		jwtSecret.Labels = maps.Clone(core.Labels)
		if jwtSecret.Labels == nil {
			jwtSecret.Labels = make(map[string]string)
		}

		jwtSecret.Labels[meta.SupabaseLabel.Reload] = ""

		if jwtSecret.Data == nil {
			jwtSecret.Data = make(map[string][]byte, 2)
		}

		var (
			plainSecret    = core.Spec.JWT.Secret
			expectedSecret = make([]byte, hex.EncodedLen(len(*plainSecret)))
			secretChanged  bool
		)

		hex.Encode(expectedSecret, []byte(*plainSecret))
		currentSecret, ok := jwtSecret.Data[core.Spec.JWT.SecretKey]
		if !ok {
			jwtSecret.Data[core.Spec.JWT.SecretKey] = expectedSecret
			secretChanged = true
		} else if !bytes.Equal(expectedSecret, currentSecret) {
			secretChanged = true
			jwtSecret.Data[core.Spec.JWT.SecretKey] = expectedSecret
		}

		core.Spec.JWT.Secret = nil

		if _, ok := jwtSecret.Data[core.Spec.JWT.JwksKey]; !ok || secretChanged {
			keySet := jwk.Set[jwk.SymmetricKey]{
				Keys: []jwk.SymmetricKey{{
					Algorithm: jwk.AlgorithmHS256,
					Key:       jwtSecret.Data[core.Spec.JWT.SecretKey],
				}},
			}

			serializedKeySet, err := json.Marshal(keySet)
			if err != nil {
				return fmt.Errorf("marshalling JWKS: %w", err)
			}

			jwtSecret.Data[core.Spec.JWT.JwksKey] = serializedKeySet
		}

		return nil
	})

	return err
}
