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
	"crypto/rand"
	"encoding/hex"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	supabasev1alpha1 "github.com/prskr/supabase-operator/api/v1alpha1"
	"github.com/prskr/supabase-operator/internal/meta"
	"github.com/prskr/supabase-operator/internal/supabase"
)

// DashboardPGMetaReconciler reconciles a Dashboard object
type DashboardPGMetaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *DashboardPGMetaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var (
		dashboard supabasev1alpha1.Dashboard
		logger    = log.FromContext(ctx)
	)

	if err := r.Get(ctx, req.NamespacedName, &dashboard); client.IgnoreNotFound(err) != nil {
		logger.Error(err, "unable to fetch Dashboard")
		return ctrl.Result{}, err
	}

	if err := r.reconcilePGMetaCryptoKeySecret(ctx, &dashboard); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcilePGMetaDeployment(ctx, &dashboard); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcilePGMetaService(ctx, &dashboard); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DashboardPGMetaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&supabasev1alpha1.Dashboard{}).
		Owns(new(appsv1.Deployment)).
		Owns(new(corev1.Service)).
		Owns(new(corev1.Secret)).
		Named("dashboard-pgmeta").
		Complete(r)
}

func (r *DashboardPGMetaReconciler) reconcilePGMetaCryptoKeySecret(
	ctx context.Context,
	dashboard *supabasev1alpha1.Dashboard,
) error {
	var (
		serviceCfg   = supabase.ServiceConfig.PGMeta
		cryptoSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceCfg.CryptoKeySecretName(dashboard),
				Namespace: dashboard.Namespace,
			},
		}
	)

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, cryptoSecret, func() error {
		if cryptoSecret.Labels == nil {
			cryptoSecret.Labels = make(map[string]string)
		}

		cryptoSecret.Labels[meta.SupabaseLabel.SecretKind] = "pg-meta-crypto-key"

		if cryptoSecret.Data == nil {
			cryptoSecret.Data = make(map[string][]byte)
		}

		if _, ok := cryptoSecret.Data[serviceCfg.Defaults.CryptoKeyKey]; !ok {
			secret := make([]byte, serviceCfg.Defaults.CryptoKeyLength)
			if _, err := rand.Read(secret); err != nil {
				return err
			}
			cryptoSecret.Data[serviceCfg.Defaults.CryptoKeyKey] = []byte(hex.EncodeToString(secret))
		}

		if err := controllerutil.SetControllerReference(dashboard, cryptoSecret, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r *DashboardPGMetaReconciler) reconcilePGMetaDeployment(
	ctx context.Context,
	dashboard *supabasev1alpha1.Dashboard,
) error {
	var (
		serviceCfg       = supabase.ServiceConfig.PGMeta
		pgMetaDeployment = &appsv1.Deployment{
			ObjectMeta: serviceCfg.ObjectMeta(dashboard),
		}
		pgMetaSpec = dashboard.Spec.PGMeta
	)

	if pgMetaSpec == nil {
		pgMetaSpec = new(supabasev1alpha1.PGMetaSpec)
	}

	dsnSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dashboard.Spec.DBSpec.DBCredentialsRef.SecretName,
			Namespace: dashboard.Namespace,
		},
	}
	if err := r.Get(ctx, client.ObjectKeyFromObject(dsnSecret), dsnSecret); err != nil {
		return err
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, pgMetaDeployment, func() error {
		pgMetaDeployment.Labels = pgMetaSpec.WorkloadSpec.MergeLabels(
			objectLabels(dashboard, serviceCfg.Name, "dashboard", supabase.Images.PostgresMeta.Tag),
			dashboard.Labels,
		)

		if pgMetaDeployment.CreationTimestamp.IsZero() {
			pgMetaDeployment.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: selectorLabels(dashboard, serviceCfg.Name),
			}
		}

		pgMetaDeployment.Spec.Replicas = pgMetaSpec.WorkloadSpec.ReplicaCount()

		pgMetaEnv := []corev1.EnvVar{
			serviceCfg.EnvKeys.APIPort.Var(serviceCfg.Defaults.APIPort),
			serviceCfg.EnvKeys.DBHost.Var(dashboard.Spec.DBSpec.Host),
			serviceCfg.EnvKeys.DBName.Var(dashboard.Spec.DBSpec.DBName),
			serviceCfg.EnvKeys.DBPort.Var(dashboard.Spec.DBSpec.Port),
			serviceCfg.EnvKeys.DBUser.Var(dashboard.Spec.DBSpec.UserRef()),
			serviceCfg.EnvKeys.DBPassword.Var(dashboard.Spec.DBSpec.PasswordRef()),
			// serviceCfg.EnvKeys.DBSSLMode.Var("require"),
			serviceCfg.EnvKeys.CryptoKey.Var(serviceCfg.CryptoKeySelector(dashboard)),
		}

		pgMetaDeployment.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: objectLabels(dashboard, serviceCfg.Name, "dashboard", supabase.Images.PostgresMeta.Tag),
			},
			Spec: corev1.PodSpec{
				ImagePullSecrets:             pgMetaSpec.WorkloadSpec.PullSecrets(),
				AutomountServiceAccountToken: new(false),
				Containers: []corev1.Container{{
					Name:            "supabase-meta",
					Image:           pgMetaSpec.WorkloadSpec.Image(supabase.Images.PostgresMeta.String()),
					ImagePullPolicy: pgMetaSpec.WorkloadSpec.ImagePullPolicy(),
					Env:             pgMetaSpec.WorkloadSpec.MergeEnv(pgMetaEnv),
					Ports: []corev1.ContainerPort{{
						Name:          "api",
						ContainerPort: serviceCfg.Defaults.APIPort,
						Protocol:      corev1.ProtocolTCP,
					}},
					SecurityContext: pgMetaSpec.WorkloadSpec.ContainerSecurityContext(serviceCfg.Defaults.NodeUID, serviceCfg.Defaults.NodeGID),
					Resources:       pgMetaSpec.WorkloadSpec.Resources(),
					VolumeMounts:    pgMetaSpec.WorkloadSpec.AdditionalVolumeMounts(),
					ReadinessProbe: &corev1.Probe{
						InitialDelaySeconds: 5,
						PeriodSeconds:       3,
						TimeoutSeconds:      1,
						SuccessThreshold:    2,
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: serviceCfg.LivenessProbePath,
								Port: intstr.IntOrString{IntVal: serviceCfg.Defaults.APIPort},
							},
						},
					},
					LivenessProbe: &corev1.Probe{
						InitialDelaySeconds: 10,
						PeriodSeconds:       5,
						TimeoutSeconds:      3,
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: serviceCfg.LivenessProbePath,
								Port: intstr.IntOrString{IntVal: serviceCfg.Defaults.APIPort},
							},
						},
					},
				}},
				SecurityContext: pgMetaSpec.WorkloadSpec.PodSecurityContext(),
			},
		}

		if err := controllerutil.SetControllerReference(dashboard, pgMetaDeployment, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r *DashboardPGMetaReconciler) reconcilePGMetaService(
	ctx context.Context,
	dashboard *supabasev1alpha1.Dashboard,
) error {
	pgMetaService := &corev1.Service{
		ObjectMeta: supabase.ServiceConfig.PGMeta.ObjectMeta(dashboard),
	}

	if dashboard.Spec.PGMeta == nil {
		dashboard.Spec.PGMeta = new(supabasev1alpha1.PGMetaSpec)
	}

	_, err := controllerutil.CreateOrPatch(ctx, r.Client, pgMetaService, func() error {
		pgMetaService.Labels = dashboard.Spec.PGMeta.WorkloadSpec.MergeLabels(
			objectLabels(dashboard, supabase.ServiceConfig.PGMeta.Name, "dashboard", supabase.Images.PostgresMeta.Tag),
			dashboard.Labels,
		)

		if _, ok := pgMetaService.Labels[meta.SupabaseLabel.ApiGatewayTarget]; !ok {
			pgMetaService.Labels[meta.SupabaseLabel.ApiGatewayTarget] = ""
		}

		pgMetaService.Spec = corev1.ServiceSpec{
			Selector: selectorLabels(dashboard, supabase.ServiceConfig.PGMeta.Name),
			Ports: []corev1.ServicePort{
				{
					Name:        "api",
					Protocol:    corev1.ProtocolTCP,
					AppProtocol: new("http"),
					Port:        supabase.ServiceConfig.PGMeta.Defaults.APIPort,
					TargetPort:  intstr.IntOrString{IntVal: supabase.ServiceConfig.PGMeta.Defaults.APIPort},
				},
			},
		}

		if err := controllerutil.SetControllerReference(dashboard, pgMetaService, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}
