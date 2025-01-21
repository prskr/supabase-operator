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
	"encoding/hex"
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

type CorePostgrestReconiler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *CorePostgrestReconiler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	var (
		core   supabasev1alpha1.Core
		logger = log.FromContext(ctx)
	)

	if err := r.Get(ctx, req.NamespacedName, &core); client.IgnoreNotFound(err) != nil {
		logger.Error(err, "unable to fetch Core")
		return ctrl.Result{}, err
	}

	if err := r.reconilePostgrestDeployment(ctx, &core); err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Error(err, "expected resource does not exist (yet), waiting for it to be present")
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
		return ctrl.Result{}, err
	}

	if err := r.reconcilePostgrestService(ctx, &core); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CorePostgrestReconiler) SetupWithManager(mgr ctrl.Manager) error {
	// TODO watch changes in DB credentials secret

	return ctrl.NewControllerManagedBy(mgr).
		For(new(supabasev1alpha1.Core)).
		Owns(new(appsv1.Deployment)).
		Owns(new(corev1.Service)).
		Named("core-postgrest").
		Complete(r)
}

func (r *CorePostgrestReconiler) reconilePostgrestDeployment(
	ctx context.Context,
	core *supabasev1alpha1.Core,
) error {
	var (
		serviceCfg          = supabase.ServiceConfig.Postgrest
		postgrestDeployment = &appsv1.Deployment{
			ObjectMeta: serviceCfg.ObjectMeta(core),
		}
		postgrestSpec = core.Spec.Postgrest
	)

	var (
		anonRole         = ValueOrFallback(postgrestSpec.AnonRole, serviceCfg.Defaults.AnonRole)
		postgrestSchemas = ValueOrFallback(postgrestSpec.Schemas, serviceCfg.Defaults.Schemas)
		jwtSecretHash    string
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

	if jwtSecret, err := core.Spec.JWT.GetJWTSecret(ctx, namespacedClient); err != nil {
		return err
	} else {
		jwtSecretHash = hex.EncodeToString(HashBytes(jwtSecret))
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, postgrestDeployment, func() error {
		postgrestDeployment.Labels = postgrestSpec.WorkloadTemplate.MergeLabels(
			objectLabels(core, serviceCfg.Name, "core", supabase.Images.Postgrest.Tag),
			core.Labels,
		)

		postgrestEnv := []corev1.EnvVar{
			{
				Name: "DB_CREDENTIALS_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: core.Spec.Database.Roles.Secrets.Authenticator,
						},
						Key: corev1.BasicAuthPasswordKey,
					},
				},
			},
			{
				Name:  serviceCfg.EnvKeys.DBUri,
				Value: strings.TrimSuffix(fmt.Sprintf("postgres://%s:$(DB_CREDENTIALS_PASSWORD)@%s%s?%s", supabase.DBRoleAuthenticator, parsedDSN.Host, parsedDSN.Path, parsedDSN.Query().Encode()), "?"),
			},
			serviceCfg.EnvKeys.Host.Var(),
			serviceCfg.EnvKeys.JWTSecret.Var(core.Spec.JWT.JwksKeySelector()),
			serviceCfg.EnvKeys.Schemas.Var(postgrestSchemas),
			serviceCfg.EnvKeys.AnonRole.Var(anonRole),
			serviceCfg.EnvKeys.UseLegacyGucs.Var(false),
			serviceCfg.EnvKeys.ExtraSearchPath.Var(serviceCfg.Defaults.ExtraSearchPath),
			serviceCfg.EnvKeys.AppSettingsJWTSecret.Var(core.Spec.JWT.SecretKeySelector()),
			serviceCfg.EnvKeys.AppSettingsJWTExpiry.Var(ValueOrFallback(core.Spec.JWT.Expiry, supabase.ServiceConfig.JWT.Defaults.Expiry)),
			serviceCfg.EnvKeys.AdminServerPort.Var((serviceCfg.Defaults.AdminPort)),
			serviceCfg.EnvKeys.MaxRows.Var(postgrestSpec.MaxRows),
			serviceCfg.EnvKeys.OpenAPIProxyURI.Var(fmt.Sprintf("%s/rest/v1", strings.TrimSuffix(core.Spec.APIExternalURL, "/"))),
		}

		if postgrestDeployment.CreationTimestamp.IsZero() {
			postgrestDeployment.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: selectorLabels(core, serviceCfg.Name),
			}
		}

		postgrestDeployment.Spec.Replicas = postgrestSpec.WorkloadTemplate.ReplicaCount()

		postgrestDeployment.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					fmt.Sprintf("%s/%s", supabasev1alpha1.GroupVersion.Group, "jwt-hash"): jwtSecretHash,
				},
				Labels: objectLabels(core, serviceCfg.Name, "core", supabase.Images.Postgrest.Tag),
			},
			Spec: corev1.PodSpec{
				ImagePullSecrets: postgrestSpec.WorkloadTemplate.PullSecrets(),
				Containers: []corev1.Container{
					{
						Name:            "supabase-rest",
						Image:           postgrestSpec.WorkloadTemplate.Image(supabase.Images.Postgrest.String()),
						ImagePullPolicy: postgrestSpec.WorkloadTemplate.ImagePullPolicy(),
						Args:            []string{"postgrest"},
						Env:             postgrestSpec.WorkloadTemplate.MergeEnv(postgrestEnv),
						Ports: []corev1.ContainerPort{
							{
								Name:          "rest",
								ContainerPort: serviceCfg.Defaults.ServerPort,
								Protocol:      corev1.ProtocolTCP,
							},
							{
								Name:          "admin",
								ContainerPort: serviceCfg.Defaults.AdminPort,
								Protocol:      corev1.ProtocolTCP,
							},
						},
						SecurityContext: postgrestSpec.WorkloadTemplate.ContainerSecurityContext(serviceCfg.Defaults.UID, serviceCfg.Defaults.GID),
						Resources:       postgrestSpec.WorkloadTemplate.Resources(),
						VolumeMounts:    postgrestSpec.WorkloadTemplate.AdditionalVolumeMounts(),
						ReadinessProbe: &corev1.Probe{
							InitialDelaySeconds: 5,
							PeriodSeconds:       3,
							TimeoutSeconds:      1,
							SuccessThreshold:    2,
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/ready",
									Port: intstr.IntOrString{IntVal: serviceCfg.Defaults.AdminPort},
								},
							},
						},
						LivenessProbe: &corev1.Probe{
							InitialDelaySeconds: 10,
							PeriodSeconds:       5,
							TimeoutSeconds:      3,
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/live",
									Port: intstr.IntOrString{IntVal: serviceCfg.Defaults.AdminPort},
								},
							},
						},
					},
				},
				SecurityContext: postgrestSpec.WorkloadTemplate.PodSecurityContext(),
			},
		}

		if err := controllerutil.SetControllerReference(core, postgrestDeployment, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r *CorePostgrestReconiler) reconcilePostgrestService(
	ctx context.Context,
	core *supabasev1alpha1.Core,
) error {
	postgrestService := &corev1.Service{
		ObjectMeta: supabase.ServiceConfig.Postgrest.ObjectMeta(core),
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, postgrestService, func() error {
		postgrestService.Labels = core.Spec.Postgrest.WorkloadTemplate.MergeLabels(
			objectLabels(core, supabase.ServiceConfig.Postgrest.Name, "core", supabase.Images.Postgrest.Tag),
			core.Labels,
		)

		if _, ok := postgrestService.Labels[meta.SupabaseLabel.ApiGatewayTarget]; !ok {
			postgrestService.Labels[meta.SupabaseLabel.ApiGatewayTarget] = ""
		}

		postgrestService.Spec = corev1.ServiceSpec{
			Selector: selectorLabels(core, supabase.ServiceConfig.Postgrest.Name),
			Ports: []corev1.ServicePort{
				{
					Name:        "rest",
					Protocol:    corev1.ProtocolTCP,
					AppProtocol: ptrOf("http"),
					Port:        3000,
					TargetPort:  intstr.IntOrString{IntVal: 3000},
				},
			},
		}

		if err := controllerutil.SetControllerReference(core, postgrestService, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}
