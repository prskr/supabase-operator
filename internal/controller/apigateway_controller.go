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
	"bytes"
	"context"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"text/template"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
	"code.icb4dc0.de/prskr/supabase-operator/internal/meta"
	"code.icb4dc0.de/prskr/supabase-operator/internal/supabase"
)

var (
	templates *template.Template
	//go:embed templates/*.tmpl
	templateFS          embed.FS
	ErrNoJwksConfigured = errors.New("no JWKS configured")
)

const (
	jwksSecretNameField = ".spec.jwks.name"
)

func init() {
	templates = template.Must(template.ParseFS(templateFS, "templates/*.tmpl"))
}

// APIGatewayReconciler reconciles a APIGateway object
type APIGatewayReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *APIGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	var (
		gateway                   supabasev1alpha1.APIGateway
		logger                    = log.FromContext(ctx)
		envoyConfigHash, jwksHash string
	)

	logger.Info("Reconciling APIGateway")

	if err := r.Get(ctx, req.NamespacedName, &gateway); client.IgnoreNotFound(err) != nil {
		logger.Error(err, "unable to fetch Gateway")
		return ctrl.Result{}, err
	}

	if jwksHash, err = r.reconcileJwksSecret(ctx, &gateway); err != nil {
		return ctrl.Result{}, err
	}

	if envoyConfigHash, err = r.reconcileEnvoyConfig(ctx, &gateway); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconileEnvoyDeployment(ctx, &gateway, envoyConfigHash, jwksHash); err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Error(err, "expected resource does not exist (yet), waiting for it to be present")
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
		return ctrl.Result{}, err
	}

	if err := r.reconcileEnvoyService(ctx, &gateway); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *APIGatewayReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	err := mgr.GetFieldIndexer().IndexField(ctx, new(supabasev1alpha1.APIGateway), jwksSecretNameField, func(o client.Object) []string {
		gw, ok := o.(*supabasev1alpha1.APIGateway)
		if !ok {
			return nil
		}

		return []string{gw.Spec.JWKSSelector.Name}
	})
	if err != nil {
		return fmt.Errorf("setting up field index for JWKS secret name: %w", err)
	}

	reloadSelector, err := predicate.LabelSelectorPredicate(metav1.LabelSelector{
		MatchLabels: map[string]string{
			meta.SupabaseLabel.Reload: "",
		},
	})
	if err != nil {
		return fmt.Errorf("constructor selector for watching secrets: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&supabasev1alpha1.APIGateway{}).
		Named("apigateway").
		Owns(new(corev1.ConfigMap)).
		Owns(new(appsv1.Deployment)).
		Owns(new(corev1.Service)).
		Watches(
			new(corev1.Secret),
			FieldSelectorEventHandler[*supabasev1alpha1.APIGateway, *supabasev1alpha1.APIGatewayList](r.Client,
				jwksSecretNameField,
			),
			builder.WithPredicates(
				predicate.ResourceVersionChangedPredicate{},
				reloadSelector,
			),
		).
		Complete(r)
}

func (r *APIGatewayReconciler) reconcileJwksSecret(
	ctx context.Context,
	gateway *supabasev1alpha1.APIGateway,
) (jwksHash string, err error) {
	jwksSecret := &corev1.Secret{ObjectMeta: gateway.JwksSecretMeta()}

	if err := r.Get(ctx, client.ObjectKeyFromObject(jwksSecret), jwksSecret); err != nil {
		return "", err
	}

	jwksRaw, ok := jwksSecret.Data[gateway.Spec.JWKSSelector.Key]
	if !ok {
		return "", fmt.Errorf("%w in secret %s", ErrNoJwksConfigured, jwksSecret.Name)
	}

	return hex.EncodeToString(HashBytes(jwksRaw)), nil
}

func (r *APIGatewayReconciler) reconcileEnvoyConfig(
	ctx context.Context,
	gateway *supabasev1alpha1.APIGateway,
) (configHash string, err error) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      supabase.ServiceConfig.Envoy.ObjectName(gateway),
			Namespace: gateway.Namespace,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, configMap, func() error {
		configMap.Labels = MergeLabels(objectLabels(gateway, "envoy", "api-gateway", supabase.Images.Postgrest.Tag), gateway.Labels)

		type nodeSpec struct {
			Cluster string
			ID      string
		}

		type controlPlaneSpec struct {
			Name string
			Host string
			Port uint16
		}

		instance := fmt.Sprintf("%s:%s", gateway.Name, gateway.Namespace)

		tmplData := struct {
			Node         nodeSpec
			ControlPlane controlPlaneSpec
		}{
			Node: nodeSpec{
				ID:      instance,
				Cluster: instance,
			},
			ControlPlane: controlPlaneSpec{
				Name: "supabase-control-plane",
				Host: gateway.Spec.Envoy.ControlPlane.Host,
				Port: gateway.Spec.Envoy.ControlPlane.Port,
			},
		}

		bytesBuf := bytes.NewBuffer(nil)

		if err := templates.ExecuteTemplate(bytesBuf, "envoy_control_plane_config.yaml.tmpl", tmplData); err != nil {
			return err
		}

		configMap.Data = map[string]string{
			"config.yaml": bytesBuf.String(),
		}

		if err := controllerutil.SetControllerReference(gateway, configMap, r.Scheme); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	configHash = hex.EncodeToString(HashStrings(configMap.Data[supabase.ServiceConfig.Envoy.Defaults.ConfigKey]))

	return configHash, nil
}

func (r *APIGatewayReconciler) reconileEnvoyDeployment(
	ctx context.Context,
	gateway *supabasev1alpha1.APIGateway,
	configHash, jwksHash string,
) error {
	envoyDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      supabase.ServiceConfig.Envoy.ObjectName(gateway),
			Namespace: gateway.Namespace,
		},
	}

	envoySpec := gateway.Spec.Envoy

	if envoySpec == nil {
		envoySpec = new(supabasev1alpha1.EnvoySpec)
	}

	if envoySpec.WorkloadTemplate == nil {
		envoySpec.WorkloadTemplate = new(supabasev1alpha1.WorkloadTemplate)
	}

	if envoySpec.WorkloadTemplate.Workload == nil {
		envoySpec.WorkloadTemplate.Workload = new(supabasev1alpha1.ContainerTemplate)
	}

	var (
		image                    = supabase.Images.Envoy.String()
		podSecurityContext       = envoySpec.WorkloadTemplate.SecurityContext
		pullPolicy               = envoySpec.WorkloadTemplate.Workload.PullPolicy
		containerSecurityContext = envoySpec.WorkloadTemplate.Workload.SecurityContext
	)

	if img := envoySpec.WorkloadTemplate.Workload.Image; img != "" {
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
			RunAsUser:                ptrOf(int64(65532)),
			RunAsGroup:               ptrOf(int64(65532)),
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

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, envoyDeployment, func() error {
		envoyDeployment.Labels = MergeLabels(
			objectLabels(gateway, "envoy", "api-gateway", supabase.Images.Postgrest.Tag),
			gateway.Labels,
			envoySpec.WorkloadTemplate.AdditionalLabels,
		)

		if envoyDeployment.CreationTimestamp.IsZero() {
			envoyDeployment.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: selectorLabels(gateway, "envoy"),
			}
		}

		envoyDeployment.Spec.Replicas = envoySpec.WorkloadTemplate.Replicas

		envoyDeployment.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					fmt.Sprintf("%s/%s", supabasev1alpha1.GroupVersion.Group, "config-hash"): configHash,
					fmt.Sprintf("%s/%s", supabasev1alpha1.GroupVersion.Group, "jwks-hash"):   jwksHash,
				},
				Labels: MergeLabels(
					objectLabels(gateway, "envoy", "api-gateway", supabase.Images.Envoy.Tag),
					envoySpec.WorkloadTemplate.AdditionalLabels,
				),
			},
			Spec: corev1.PodSpec{
				ImagePullSecrets:             envoySpec.WorkloadTemplate.Workload.ImagePullSecrets,
				AutomountServiceAccountToken: ptrOf(false),
				Containers: []corev1.Container{
					{
						Name:            "envoy-proxy",
						Image:           image,
						ImagePullPolicy: pullPolicy,
						Args:            []string{"-c /etc/envoy/config.yaml"},
						Ports: []corev1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: 8000,
								Protocol:      corev1.ProtocolTCP,
							},
							{
								Name:          "admin",
								ContainerPort: 19000,
								Protocol:      corev1.ProtocolTCP,
							},
						},
						ReadinessProbe: &corev1.Probe{
							InitialDelaySeconds: 5,
							PeriodSeconds:       3,
							TimeoutSeconds:      1,
							SuccessThreshold:    2,
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/ready",
									Port: intstr.IntOrString{IntVal: 19000},
								},
							},
						},
						LivenessProbe: &corev1.Probe{
							InitialDelaySeconds: 10,
							PeriodSeconds:       5,
							TimeoutSeconds:      3,
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/ready",
									Port: intstr.IntOrString{IntVal: 19000},
								},
							},
						},
						SecurityContext: containerSecurityContext,
						Resources:       envoySpec.WorkloadTemplate.Workload.Resources,
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "config",
								ReadOnly:  true,
								MountPath: "/etc/envoy",
							},
						},
					},
				},
				SecurityContext: podSecurityContext,
				Volumes: []corev1.Volume{
					{
						Name: "config",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								Sources: []corev1.VolumeProjection{
									{
										ConfigMap: &corev1.ConfigMapProjection{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: supabase.ServiceConfig.Envoy.ObjectName(gateway),
											},
											Items: []corev1.KeyToPath{{
												Key:  "config.yaml",
												Path: "config.yaml",
											}},
										},
									},
									{
										Secret: &corev1.SecretProjection{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: gateway.Spec.JWKSSelector.Name,
											},
											Items: []corev1.KeyToPath{{
												Key:  gateway.Spec.JWKSSelector.Key,
												Path: "jwks.json",
											}},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		if err := controllerutil.SetControllerReference(gateway, envoyDeployment, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r *APIGatewayReconciler) reconcileEnvoyService(
	ctx context.Context,
	gateway *supabasev1alpha1.APIGateway,
) error {
	envoyService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      supabase.ServiceConfig.Envoy.ObjectName(gateway),
			Namespace: gateway.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, envoyService, func() error {
		envoyService.Labels = MergeLabels(objectLabels(gateway, "envoy", "api-gateway", supabase.Images.Postgrest.Tag), gateway.Labels)

		envoyService.Spec = corev1.ServiceSpec{
			Selector: selectorLabels(gateway, "postgrest"),
			Ports: []corev1.ServicePort{
				{
					Name:        "rest",
					Protocol:    corev1.ProtocolTCP,
					AppProtocol: ptrOf("http"),
					Port:        8000,
					TargetPort:  intstr.IntOrString{IntVal: 8000},
				},
			},
		}

		if err := controllerutil.SetControllerReference(gateway, envoyService, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}
