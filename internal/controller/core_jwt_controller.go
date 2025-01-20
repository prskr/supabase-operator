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
	"encoding/json"
	"fmt"
	"maps"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
	"code.icb4dc0.de/prskr/supabase-operator/internal/jwk"
	"code.icb4dc0.de/prskr/supabase-operator/internal/meta"
	"code.icb4dc0.de/prskr/supabase-operator/internal/supabase"
)

// CoreDbReconciler reconciles a Core object
type CoreJwtReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *CoreJwtReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	var (
		core   supabasev1alpha1.Core
		logger = log.FromContext(ctx)
	)

	if err := r.Get(ctx, req.NamespacedName, &core); client.IgnoreNotFound(err) != nil {
		logger.Error(err, "unable to fetch Core")
		return ctrl.Result{}, err
	}

	jwtSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: core.Spec.JWT.SecretRef.Name, Namespace: core.Namespace},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, jwtSecret, func() error {
		const (
			secretJwksAndDefaultJWTs = 4
		)

		var modifiedSecret bool

		jwtSecret.Labels = maps.Clone(core.Labels)
		if jwtSecret.Labels == nil {
			jwtSecret.Labels = make(map[string]string)
		}

		jwtSecret.Labels[meta.SupabaseLabel.Reload] = ""

		if err := controllerutil.SetControllerReference(&core, jwtSecret, r.Scheme); err != nil {
			return err
		}

		if jwtSecret.Data == nil {
			jwtSecret.Data = make(map[string][]byte, secretJwksAndDefaultJWTs)
		}

		// if secret does not contain the JWT secret as configured
		if value := jwtSecret.Data[core.Spec.JWT.SecretKey]; len(value) == 0 {
			logger.Info("Generating new JWT secret")
			generatedSecret, err := supabase.RandomJWTSecret()
			if err != nil {
				return err
			}

			jwtSecret.Data[core.Spec.JWT.SecretKey] = generatedSecret
			modifiedSecret = true
		}

		if value := jwtSecret.Data[core.Spec.JWT.JwksKey]; len(value) == 0 || modifiedSecret {
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

		if value := jwtSecret.Data[core.Spec.JWT.AnonKey]; len(value) == 0 || modifiedSecret {
			anonKey, err := generateJwt("anon", jwtSecret.Data[core.Spec.JWT.SecretKey])
			if err != nil {
				return err
			}

			jwtSecret.Data[core.Spec.JWT.AnonKey] = anonKey
		}

		if value := jwtSecret.Data[core.Spec.JWT.ServiceKey]; len(value) == 0 || modifiedSecret {
			serviceKey, err := generateJwt("service_role", jwtSecret.Data[core.Spec.JWT.SecretKey])
			if err != nil {
				return err
			}

			jwtSecret.Data[core.Spec.JWT.ServiceKey] = serviceKey
		}

		return nil
	})
	if err != nil {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CoreJwtReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(new(supabasev1alpha1.Core)).
		Owns(new(corev1.Secret)).
		Named("core-jwt").
		Complete(r)
}

func generateJwt(role string, secret []byte) ([]byte, error) {
	claims := map[string]any{
		"role":            role,
		jwt.IssuerKey:     "supabase",
		jwt.IssuedAtKey:   time.Now().Add(-30 * time.Second),
		jwt.ExpirationKey: time.Now().Add(365 * 24 * time.Hour),
	}

	token := jwt.New()

	for k, v := range claims {
		if err := token.Set(k, v); err != nil {
			return nil, err
		}
	}

	return jwt.Sign(token, jwt.WithKey(jwa.HS256, secret))
}
