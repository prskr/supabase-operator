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
		authSpec = core.Spec.Auth
		svcCfg   = supabase.ServiceConfig.Auth
	)

	if authSpec.WorkloadTemplate == nil {
		authSpec.WorkloadTemplate = new(supabasev1alpha1.WorkloadTemplate)
	}

	if authSpec.WorkloadTemplate.Workload == nil {
		authSpec.WorkloadTemplate.Workload = new(supabasev1alpha1.ContainerTemplate)
	}

	var (
		image                    = supabase.Images.Gotrue.String()
		podSecurityContext       = authSpec.WorkloadTemplate.SecurityContext
		pullPolicy               = authSpec.WorkloadTemplate.Workload.PullPolicy
		containerSecurityContext = authSpec.WorkloadTemplate.Workload.SecurityContext
		namespacedClient         = client.NewNamespacedClient(r.Client, core.Namespace)
	)

	if img := authSpec.WorkloadTemplate.Workload.Image; img != "" {
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

	databaseDSN, err := core.Spec.Database.GetDSN(ctx, namespacedClient)
	if err != nil {
		return err
	}

	parsedDSN, err := url.Parse(databaseDSN)
	if err != nil {
		return fmt.Errorf("failed to parse DB DSN: %w", err)
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, authDeployment, func() error {
		authDeployment.Labels = MergeLabels(
			objectLabels(core, "auth", "core", supabase.Images.Gotrue.Tag),
			core.Labels,
		)

		authDbEnv := []corev1.EnvVar{
			{
				Name: "POSTGRES_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: core.Spec.Database.Roles.Secrets.AuthAdmin.Name,
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
			svcCfg.EnvKeys.ApiHost.Var(svcCfg.Defaults.ApiHost),
			svcCfg.EnvKeys.ApiPort.Var(svcCfg.Defaults.ApiPort),
			svcCfg.EnvKeys.ApiExternalUrl.Var(authSpec.APIExternalURL),
			svcCfg.EnvKeys.DBDriver.Var(svcCfg.Defaults.DbDriver),
			svcCfg.EnvKeys.SiteUrl.Var(authSpec.SiteURL),
			svcCfg.EnvKeys.AdditionalRedirectURLs.Var(authSpec.AdditionalRedirectUrls),
			svcCfg.EnvKeys.DisableSignup.Var(boolValueOf(authSpec.DisableSignup)),
			svcCfg.EnvKeys.JWTIssuer.Var(svcCfg.Defaults.JwtIssuer),
			svcCfg.EnvKeys.JWTAdminRoles.Var(svcCfg.Defaults.JwtAdminRoles),
			svcCfg.EnvKeys.JWTAudience.Var(svcCfg.Defaults.JwtAudience),
			svcCfg.EnvKeys.JwtDefaultGroup.Var(svcCfg.Defaults.JwtDefaultGroupName),
			svcCfg.EnvKeys.JwtExpiry.Var(ValueOrFallback(core.Spec.JWT.Expiry, supabase.ServiceConfig.JWT.Defaults.Expiry)),
			svcCfg.EnvKeys.JwtSecret.Var(core.Spec.JWT.SecretKeySelector()),
			svcCfg.EnvKeys.EmailSignupDisabled.Var(boolValueOf(authSpec.EmailSignupDisabled)),
			svcCfg.EnvKeys.AnonymousUsersEnabled.Var(boolValueOf(authSpec.AnonymousUsersEnabled)),
		)

		authEnv = append(authEnv, authSpec.Providers.Vars(authSpec.APIExternalURL)...)

		if authDeployment.CreationTimestamp.IsZero() {
			authDeployment.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: selectorLabels(core, "auth"),
			}
		}

		authDeployment.Spec.Replicas = authSpec.WorkloadTemplate.Replicas

		authDeployment.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: objectLabels(core, "auth", "core", supabase.Images.Gotrue.Tag),
			},
			Spec: corev1.PodSpec{
				ImagePullSecrets: authSpec.WorkloadTemplate.Workload.ImagePullSecrets,
				InitContainers: []corev1.Container{{
					Name:            "migrations",
					Image:           image,
					ImagePullPolicy: pullPolicy,
					Command:         []string{"/usr/local/bin/auth"},
					Args:            []string{"migrate"},
					Env:             authEnv,
					SecurityContext: containerSecurityContext,
				}},
				Containers: []corev1.Container{{
					Name:            "supabase-auth",
					Image:           image,
					ImagePullPolicy: pullPolicy,
					Command:         []string{"/usr/local/bin/auth"},
					Args:            []string{"serve"},
					Env:             MergeEnv(authEnv, authSpec.WorkloadTemplate.Workload.AdditionalEnv...),
					Ports: []corev1.ContainerPort{{
						Name:          "api",
						ContainerPort: 9999,
						Protocol:      corev1.ProtocolTCP,
					}},
					SecurityContext: containerSecurityContext,
					Resources:       authSpec.WorkloadTemplate.Workload.Resources,
					VolumeMounts:    authSpec.WorkloadTemplate.Workload.VolumeMounts,
					ReadinessProbe: &corev1.Probe{
						InitialDelaySeconds: 5,
						PeriodSeconds:       3,
						TimeoutSeconds:      1,
						SuccessThreshold:    2,
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/health",
								Port: intstr.IntOrString{IntVal: 9999},
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
								Port: intstr.IntOrString{IntVal: 9999},
							},
						},
					},
				}},
				SecurityContext: podSecurityContext,
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
		authService.Labels = MergeLabels(
			objectLabels(core, "auth", "core", supabase.Images.Gotrue.Tag),
			core.Labels,
		)

		authService.Labels[meta.SupabaseLabel.EnvoyCluster] = core.Name

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
