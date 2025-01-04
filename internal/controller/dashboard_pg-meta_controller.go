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

package controller

import (
	"context"

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

// DashboardPGMetaReconciler reconciles a Dashboard object
type DashboardPGMetaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=dashboards,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=dashboards/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=dashboards/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Dashboard object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *DashboardPGMetaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var (
		dashboard supabasev1alpha1.Dashboard
		logger    = log.FromContext(ctx)
	)

	if err := r.Get(ctx, req.NamespacedName, &dashboard); client.IgnoreNotFound(err) != nil {
		logger.Error(err, "unable to fetch Dashboard")
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
		Named("dashboard-pgmeta").
		Complete(r)
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

	if pgMetaSpec.WorkloadTemplate == nil {
		pgMetaSpec.WorkloadTemplate = new(supabasev1alpha1.WorkloadTemplate)
	}

	if pgMetaSpec.WorkloadTemplate.Workload == nil {
		pgMetaSpec.WorkloadTemplate.Workload = new(supabasev1alpha1.ContainerTemplate)
	}

	var (
		image                    = supabase.Images.PostgresMeta.String()
		podSecurityContext       = pgMetaSpec.WorkloadTemplate.SecurityContext
		pullPolicy               = pgMetaSpec.WorkloadTemplate.Workload.PullPolicy
		containerSecurityContext = pgMetaSpec.WorkloadTemplate.Workload.SecurityContext
	)

	if img := pgMetaSpec.WorkloadTemplate.Workload.Image; img != "" {
		image = img
	}

	if podSecurityContext == nil {
		podSecurityContext = &corev1.PodSecurityContext{
			RunAsNonRoot: ptrOf(true),
		}
	}

	if containerSecurityContext == nil {
		containerSecurityContext = &corev1.SecurityContext{
			Privileged:               ptrOf(false),
			RunAsUser:                ptrOf(int64(1000)),
			RunAsGroup:               ptrOf(int64(1000)),
			RunAsNonRoot:             ptrOf(true),
			AllowPrivilegeEscalation: ptrOf(false),
			ReadOnlyRootFilesystem:   ptrOf(true),
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{
					"ALL",
				},
			},
		}
	}

	dsnSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dashboard.Spec.DBSpec.DBCredentialsRef.Name,
			Namespace: dashboard.Namespace,
		},
	}
	if err := r.Get(ctx, client.ObjectKeyFromObject(dsnSecret), dsnSecret); err != nil {
		return err
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, pgMetaDeployment, func() error {
		pgMetaDeployment.Labels = MergeLabels(
			objectLabels(dashboard, serviceCfg.Name, "dashboard", supabase.Images.PostgresMeta.Tag),
			dashboard.Labels,
		)

		if pgMetaDeployment.CreationTimestamp.IsZero() {
			pgMetaDeployment.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: selectorLabels(dashboard, serviceCfg.Name),
			}
		}

		pgMetaDeployment.Spec.Replicas = pgMetaSpec.WorkloadTemplate.Replicas

		pgMetaEnv := []corev1.EnvVar{
			serviceCfg.EnvKeys.APIPort.Var(serviceCfg.Defaults.APIPort),
			serviceCfg.EnvKeys.DBHost.Var(dashboard.Spec.DBSpec.Host),
			serviceCfg.EnvKeys.DBPort.Var(dashboard.Spec.DBSpec.Port),
			serviceCfg.EnvKeys.DBUser.Var(dashboard.Spec.DBSpec.UserRef()),
			serviceCfg.EnvKeys.DBPassword.Var(dashboard.Spec.DBSpec.PasswordRef()),
		}

		pgMetaDeployment.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: objectLabels(dashboard, serviceCfg.Name, "dashboard", supabase.Images.PostgresMeta.Tag),
			},
			Spec: corev1.PodSpec{
				ImagePullSecrets: pgMetaSpec.WorkloadTemplate.Workload.ImagePullSecrets,
				Containers: []corev1.Container{{
					Name:            "supabase-meta",
					Image:           image,
					ImagePullPolicy: pullPolicy,
					Env:             MergeEnv(pgMetaEnv, pgMetaSpec.WorkloadTemplate.Workload.AdditionalEnv...),
					Ports: []corev1.ContainerPort{{
						Name:          "api",
						ContainerPort: int32(serviceCfg.Defaults.APIPort),
						Protocol:      corev1.ProtocolTCP,
					}},
					SecurityContext: containerSecurityContext,
					Resources:       pgMetaSpec.WorkloadTemplate.Workload.Resources,
					VolumeMounts:    pgMetaSpec.WorkloadTemplate.Workload.VolumeMounts,
					ReadinessProbe: &corev1.Probe{
						InitialDelaySeconds: 5,
						PeriodSeconds:       3,
						TimeoutSeconds:      1,
						SuccessThreshold:    2,
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/health",
								Port: intstr.IntOrString{IntVal: int32(serviceCfg.Defaults.APIPort)},
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
								Port: intstr.IntOrString{IntVal: int32(serviceCfg.Defaults.APIPort)},
							},
						},
					},
				}},
				SecurityContext: podSecurityContext,
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

	_, err := controllerutil.CreateOrPatch(ctx, r.Client, pgMetaService, func() error {
		pgMetaService.Labels = MergeLabels(
			objectLabels(dashboard, supabase.ServiceConfig.PGMeta.Name, "dashboard", supabase.Images.PostgresMeta.Tag),
			dashboard.Labels,
		)

		pgMetaService.Labels[meta.SupabaseLabel.EnvoyCluster] = dashboard.Name

		apiPort := int32(supabase.ServiceConfig.PGMeta.Defaults.APIPort)

		pgMetaService.Spec = corev1.ServiceSpec{
			Selector: selectorLabels(dashboard, supabase.ServiceConfig.PGMeta.Name),
			Ports: []corev1.ServicePort{
				{
					Name:        "api",
					Protocol:    corev1.ProtocolTCP,
					AppProtocol: ptrOf("http"),
					Port:        apiPort,
					TargetPort:  intstr.IntOrString{IntVal: apiPort},
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
