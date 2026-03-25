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
	"errors"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	supabasev1alpha1 "github.com/prskr/supabase-operator/api/v1alpha1"
	"github.com/prskr/supabase-operator/internal/meta"
	"github.com/prskr/supabase-operator/internal/supabase"
)

var errAmbiguousGatewayService = errors.New("ambiguous gateway service matches")

type DashboardStudioReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder events.EventRecorder
}

func (r *DashboardStudioReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var (
		dashboard supabasev1alpha1.Dashboard
		logger    = log.FromContext(ctx)
	)

	if err := r.Get(ctx, req.NamespacedName, &dashboard); client.IgnoreNotFound(err) != nil {
		logger.Error(err, "unable to fetch Dashboard")
		return ctrl.Result{}, err
	}

	if err := r.reconcileStudioDeployment(ctx, &dashboard); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileStudioService(ctx, &dashboard); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DashboardStudioReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&supabasev1alpha1.Dashboard{}).
		Owns(new(appsv1.Deployment)).
		Owns(new(corev1.Service)).
		Owns(new(corev1.Secret)).
		Named("dashboard-studio").
		Complete(r)
}

func (r *DashboardStudioReconciler) reconcileStudioDeployment(
	ctx context.Context,
	dashboard *supabasev1alpha1.Dashboard,
) error {
	var (
		serviceCfg       = supabase.ServiceConfig.Studio
		studioDeployment = &appsv1.Deployment{
			ObjectMeta: serviceCfg.ObjectMeta(dashboard),
		}
		studioSpec             = dashboard.Spec.Studio
		gatewayServiceList     corev1.ServiceList
		dbSchemasValue         = ValueOrFallback(dashboard.Spec.DBSpec.Schemas, serviceCfg.Defaults.Schemas)
		extraDBSearchPathValue = ValueOrFallback(dashboard.Spec.DBSpec.ExtraSearchPath, serviceCfg.Defaults.ExtraSearchPath)
	)

	if studioSpec == nil {
		studioSpec = new(supabasev1alpha1.StudioSpec)
	}

	err := r.List(
		ctx,
		&gatewayServiceList,
		client.InNamespace(dashboard.Namespace),
		client.MatchingLabels(studioSpec.GatewayServiceMatchLabels),
	)
	if err != nil {
		return fmt.Errorf("selecting gateway service: %w", err)
	}

	if itemCount := len(gatewayServiceList.Items); itemCount < 1 {
		r.Recorder.Eventf(
			dashboard,
			dashboard,
			corev1.EventTypeWarning,
			"GatewayServiceNotFound",
			"Cannot configure link from Studio to API Gateway",
			"Make sure that the API Gateway is deployed and a matching selector is configured in order to get a functioning Studio deployment",
		)

		return fmt.Errorf("%w: no matching service found", errAmbiguousGatewayService)
	} else if itemCount > 1 {
		r.Recorder.Eventf(
			dashboard,
			dashboard,
			corev1.EventTypeWarning,
			"AmbiguousGatewayServicesFound",
			"Cannot configure link from Studio to API Gateway",
			"Make sure that the API Gateway is deployed and a matching selector is configured in order to get a functioning Studio deployment",
		)

		return fmt.Errorf("%w: multiple ( %d ) matching services found", errAmbiguousGatewayService, itemCount)
	}

	gatewayService := gatewayServiceList.Items[0]

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, studioDeployment, func() error {
		studioDeployment.Labels = studioSpec.WorkloadSpec.MergeLabels(
			objectLabels(dashboard, serviceCfg.Name, "dashboard", supabase.Images.Studio.Tag),
			dashboard.Labels,
		)

		if studioDeployment.CreationTimestamp.IsZero() {
			studioDeployment.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: selectorLabels(dashboard, serviceCfg.Name),
			}
		}

		studioDeployment.Spec.Replicas = studioSpec.WorkloadSpec.ReplicaCount()

		studioEnv := []corev1.EnvVar{
			serviceCfg.EnvKeys.PGMetaURL.Var(fmt.Sprintf("http://%s.%s.svc:%d", supabase.ServiceConfig.PGMeta.ObjectName(dashboard), dashboard.Namespace, supabase.ServiceConfig.PGMeta.Defaults.APIPort)),
			serviceCfg.EnvKeys.Host.Var(),
			serviceCfg.EnvKeys.ApiUrl.Var(fmt.Sprintf("http://%s.%s.svc:8000", gatewayService.Name, gatewayService.Namespace)),
			serviceCfg.EnvKeys.DBHost.Var(dashboard.Spec.DBSpec.Host),
			serviceCfg.EnvKeys.DBPort.Var(dashboard.Spec.DBSpec.Port),
			serviceCfg.EnvKeys.DBName.Var(dashboard.Spec.DBSpec.DBName),
			serviceCfg.EnvKeys.DBPassword.Var(dashboard.Spec.DBSpec.PasswordRef()),
			serviceCfg.EnvKeys.DBSchemas.Var(dbSchemasValue),
			serviceCfg.EnvKeys.DBExtraSearchPath.Var(extraDBSearchPathValue),
			serviceCfg.EnvKeys.DBMaxRows.Var(serviceCfg.Defaults.MaxRows),
			serviceCfg.EnvKeys.APIExternalURL.Var(studioSpec.APIExternalURL),
			serviceCfg.EnvKeys.JwtSecret.Var(studioSpec.JWT.SecretKeySelector()),
			serviceCfg.EnvKeys.AnonKey.Var(studioSpec.JWT.AnonKeySelector()),
			serviceCfg.EnvKeys.ServiceKey.Var(studioSpec.JWT.ServiceKeySelector()),
			serviceCfg.EnvKeys.LogsEnabled.Var(),
			serviceCfg.EnvKeys.MetaCryptoKey.Var(supabase.ServiceConfig.PGMeta.CryptoKeySelector(dashboard)),
		}

		if studioSpec.AI != nil {
			studioEnv = append(studioEnv, studioSpec.AI.OpenAI.Vars()...)
		}

		studioDeployment.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: objectLabels(dashboard, serviceCfg.Name, "dashboard", supabase.Images.Studio.Tag),
			},
			Spec: corev1.PodSpec{
				ImagePullSecrets:             studioSpec.WorkloadSpec.PullSecrets(),
				AutomountServiceAccountToken: new(false),
				Containers: []corev1.Container{{
					Name:            "supabase-studio",
					Image:           studioSpec.WorkloadSpec.Image(supabase.Images.Studio.String()),
					ImagePullPolicy: studioSpec.WorkloadSpec.ImagePullPolicy(),
					Env:             studioSpec.WorkloadSpec.MergeEnv(studioEnv),
					Ports: []corev1.ContainerPort{{
						Name:          "studio",
						ContainerPort: serviceCfg.Defaults.APIPort,
						Protocol:      corev1.ProtocolTCP,
					}},
					SecurityContext: studioSpec.WorkloadSpec.ContainerSecurityContext(serviceCfg.Defaults.NodeUID, serviceCfg.Defaults.NodeGID),
					Resources:       studioSpec.WorkloadSpec.Resources(),
					VolumeMounts: studioSpec.WorkloadSpec.AdditionalVolumeMounts(corev1.VolumeMount{
						Name:      "next-cache",
						MountPath: "/app/apps/studio/.next/cache",
					}),
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
				SecurityContext: studioSpec.WorkloadSpec.PodSecurityContext(),
				Volumes: []corev1.Volume{{
					Name: "next-cache",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium:    corev1.StorageMediumDefault,
							SizeLimit: new(resource.MustParse("300Mi")),
						},
					},
				}},
			},
		}

		if err := controllerutil.SetControllerReference(dashboard, studioDeployment, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r *DashboardStudioReconciler) reconcileStudioService(
	ctx context.Context,
	dashboard *supabasev1alpha1.Dashboard,
) error {
	studioService := &corev1.Service{
		ObjectMeta: supabase.ServiceConfig.Studio.ObjectMeta(dashboard),
	}

	if dashboard.Spec.Studio == nil {
		dashboard.Spec.Studio = new(supabasev1alpha1.StudioSpec)
	}

	_, err := controllerutil.CreateOrPatch(ctx, r.Client, studioService, func() error {
		studioService.Labels = dashboard.Spec.Studio.WorkloadSpec.MergeLabels(
			objectLabels(dashboard, supabase.ServiceConfig.Studio.Name, "dashboard", supabase.Images.Studio.Tag),
			dashboard.Labels,
		)

		if _, ok := studioService.Labels[meta.SupabaseLabel.ApiGatewayTarget]; !ok {
			studioService.Labels[meta.SupabaseLabel.ApiGatewayTarget] = ""
		}

		studioService.Spec = corev1.ServiceSpec{
			Selector: selectorLabels(dashboard, supabase.ServiceConfig.Studio.Name),
			Ports: []corev1.ServicePort{
				{
					Name:        "studio",
					Protocol:    corev1.ProtocolTCP,
					AppProtocol: new("http"),
					Port:        supabase.ServiceConfig.Studio.Defaults.APIPort,
					TargetPort:  intstr.IntOrString{IntVal: supabase.ServiceConfig.Studio.Defaults.APIPort},
				},
			},
		}

		if err := controllerutil.SetControllerReference(dashboard, studioService, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}
