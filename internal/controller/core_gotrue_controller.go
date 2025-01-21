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
	"fmt"
	"net/url"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
	"code.icb4dc0.de/prskr/supabase-operator/internal/meta"
	"code.icb4dc0.de/prskr/supabase-operator/internal/supabase"
)

type CoreAuthReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *CoreAuthReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	var (
		core   supabasev1alpha1.Core
		logger = log.FromContext(ctx)
	)

	if err := r.Get(ctx, req.NamespacedName, &core); client.IgnoreNotFound(err) != nil {
		logger.Error(err, "unable to fetch Core")
		return ctrl.Result{}, err
	}

	if err := r.reconcileAuthDeployment(ctx, &core); err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Error(err, "expected resource does not exist (yet), waiting for it to be present")
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
		return ctrl.Result{}, err
	}

	if err := r.reconcileAuthService(ctx, &core); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CoreAuthReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// TODO watch changes in DB credentials secret
	return ctrl.NewControllerManagedBy(mgr).
		For(new(supabasev1alpha1.Core)).
		Owns(new(appsv1.Deployment)).
		Owns(new(corev1.Service)).
		Named("core-auth").
		Complete(r)
}

func (r *CoreAuthReconciler) reconcileAuthDeployment(
	ctx context.Context,
	core *supabasev1alpha1.Core,
) error {
	var (
		authDeployment = &appsv1.Deployment{
			ObjectMeta: supabase.ServiceConfig.Auth.ObjectMeta(core),
		}
		authSpec         = core.Spec.Auth
		svcCfg           = supabase.ServiceConfig.Auth
		namespacedClient = client.NewNamespacedClient(r.Client, core.Namespace)
	)

	databaseDSN, err := core.Spec.Database.GetDSN(ctx, namespacedClient)
	if err != nil {
		return err
	}

	parsedDSN, err := url.Parse(databaseDSN)
	if err != nil {
		return fmt.Errorf("failed to parse DB DSN: %w", err)
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, authDeployment, func() error {
		authDeployment.Labels = authSpec.WorkloadTemplate.MergeLabels(
			objectLabels(core, "auth", "core", supabase.Images.Gotrue.Tag),
			core.Labels,
		)

		authDbEnv := []corev1.EnvVar{
			{
				Name: "POSTGRES_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: core.Spec.Database.Roles.Secrets.AuthAdmin,
						},
						Key: corev1.BasicAuthPasswordKey,
					},
				},
			},
			{
				Name:  svcCfg.EnvKeys.DatabaseUrl,
				Value: strings.TrimSuffix(fmt.Sprintf("postgres://%s:$(POSTGRES_PASSWORD)@%s%s?%s", supabase.DBRoleAuthAdmin, parsedDSN.Host, parsedDSN.Path, parsedDSN.Query().Encode()), "?"),
			},
		}

		authEnv := append(authDbEnv,
			svcCfg.EnvKeys.ApiHost.Var(),
			svcCfg.EnvKeys.ApiPort.Var(),
			svcCfg.EnvKeys.ApiExternalUrl.Var(core.Spec.APIExternalURL),
			svcCfg.EnvKeys.DBDriver.Var(),
			svcCfg.EnvKeys.SiteUrl.Var(core.Spec.SiteURL),
			svcCfg.EnvKeys.AdditionalRedirectURLs.Var(authSpec.AdditionalRedirectUrls),
			svcCfg.EnvKeys.DisableSignup.Var(boolValueOf(authSpec.DisableSignup)),
			svcCfg.EnvKeys.JWTIssuer.Var(),
			svcCfg.EnvKeys.JWTAdminRoles.Var(),
			svcCfg.EnvKeys.JWTAudience.Var(),
			svcCfg.EnvKeys.JwtDefaultGroup.Var(),
			svcCfg.EnvKeys.JwtExpiry.Var(ValueOrFallback(core.Spec.JWT.Expiry, supabase.ServiceConfig.JWT.Defaults.Expiry)),
			svcCfg.EnvKeys.JwtSecret.Var(core.Spec.JWT.SecretKeySelector()),
			svcCfg.EnvKeys.EmailSignupDisabled.Var(boolValueOf(authSpec.EmailSignupDisabled)),
			svcCfg.EnvKeys.AnonymousUsersEnabled.Var(boolValueOf(authSpec.AnonymousUsersEnabled)),
		)

		authEnv = append(authEnv, authSpec.Providers.Vars(core.Spec.APIExternalURL)...)

		if authDeployment.CreationTimestamp.IsZero() {
			authDeployment.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: selectorLabels(core, "auth"),
			}
		}

		authDeployment.Spec.Replicas = authSpec.WorkloadTemplate.ReplicaCount()

		authDeployment.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: objectLabels(core, "auth", "core", supabase.Images.Gotrue.Tag),
			},
			Spec: corev1.PodSpec{
				ImagePullSecrets: authSpec.WorkloadTemplate.PullSecrets(),
				InitContainers: []corev1.Container{{
					Name:            "supabase-auth-migrations",
					Image:           authSpec.WorkloadTemplate.Image(supabase.Images.Gotrue.String()),
					ImagePullPolicy: authSpec.WorkloadTemplate.ImagePullPolicy(),
					Command:         []string{"/usr/local/bin/auth"},
					Args:            []string{"migrate"},
					Env:             authSpec.WorkloadTemplate.MergeEnv(authEnv),
					SecurityContext: authSpec.WorkloadTemplate.ContainerSecurityContext(svcCfg.Defaults.UID, svcCfg.Defaults.GID),
				}},
				Containers: []corev1.Container{{
					Name:            "supabase-auth",
					Image:           authSpec.WorkloadTemplate.Image(supabase.Images.Gotrue.String()),
					ImagePullPolicy: authSpec.WorkloadTemplate.ImagePullPolicy(),
					Command:         []string{"/usr/local/bin/auth"},
					Args:            []string{"serve"},
					Env:             authSpec.WorkloadTemplate.MergeEnv(authEnv),
					Ports: []corev1.ContainerPort{{
						Name:          "api",
						ContainerPort: svcCfg.Defaults.APIPort,
						Protocol:      corev1.ProtocolTCP,
					}},
					SecurityContext: authSpec.WorkloadTemplate.ContainerSecurityContext(svcCfg.Defaults.UID, svcCfg.Defaults.GID),
					Resources:       authSpec.WorkloadTemplate.Resources(),
					VolumeMounts:    authSpec.WorkloadTemplate.AdditionalVolumeMounts(),
					ReadinessProbe: &corev1.Probe{
						InitialDelaySeconds: 5,
						PeriodSeconds:       3,
						TimeoutSeconds:      1,
						SuccessThreshold:    2,
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/health",
								Port: intstr.IntOrString{IntVal: svcCfg.Defaults.APIPort},
							},
						},
					},
					LivenessProbe: &corev1.Probe{
						InitialDelaySeconds: 10,
						PeriodSeconds:       5,
						TimeoutSeconds:      3,
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/health",
								Port: intstr.IntOrString{IntVal: svcCfg.Defaults.APIPort},
							},
						},
					},
				}},
				SecurityContext: authSpec.WorkloadTemplate.PodSecurityContext(),
			},
		}

		if err := controllerutil.SetControllerReference(core, authDeployment, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r *CoreAuthReconciler) reconcileAuthService(
	ctx context.Context,
	core *supabasev1alpha1.Core,
) error {
	authService := &corev1.Service{
		ObjectMeta: supabase.ServiceConfig.Auth.ObjectMeta(core),
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, authService, func() error {
		authService.Labels = core.Spec.Postgrest.WorkloadTemplate.MergeLabels(
			objectLabels(core, "auth", "core", supabase.Images.Gotrue.Tag),
			core.Labels,
		)

		if _, ok := authService.Labels[meta.SupabaseLabel.ApiGatewayTarget]; !ok {
			authService.Labels[meta.SupabaseLabel.ApiGatewayTarget] = core.Name
		}

		authService.Spec = corev1.ServiceSpec{
			Selector: selectorLabels(core, "auth"),
			Ports: []corev1.ServicePort{
				{
					Name:        "api",
					Protocol:    corev1.ProtocolTCP,
					AppProtocol: ptrOf("http"),
					Port:        9999,
					TargetPort:  intstr.IntOrString{IntVal: 9999},
				},
			},
		}

		if err := controllerutil.SetControllerReference(core, authService, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}
