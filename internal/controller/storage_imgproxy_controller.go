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
	"code.icb4dc0.de/prskr/supabase-operator/internal/supabase"
)

type StorageImgProxyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.4/pkg/reconcile
func (r *StorageImgProxyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var (
		storage supabasev1alpha1.Storage
		logger  = log.FromContext(ctx)
	)

	if err := r.Get(ctx, req.NamespacedName, &storage); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	logger.Info("Reconciling Storage API")

	if storage.Spec.ImageProxy == nil || !storage.Spec.ImageProxy.Enable {
		logger.Info("ImgProxy is not enabled - skipping")
		return ctrl.Result{}, nil
	}

	if err := r.reconcileImgProxyDeployment(ctx, &storage); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileImgProxyService(ctx, &storage); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StorageImgProxyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&supabasev1alpha1.Storage{}).
		Named("storage-imgproxy").
		Owns(new(appsv1.Deployment)).
		Owns(new(corev1.Service)).
		Complete(r)
}

func (r *StorageImgProxyReconciler) reconcileImgProxyDeployment(
	ctx context.Context,
	storage *supabasev1alpha1.Storage,
) error {
	var (
		serviceCfg         = supabase.ServiceConfig.ImgProxy
		imgProxySpec       = storage.Spec.ImageProxy
		imgProxyDeployment = &appsv1.Deployment{
			ObjectMeta: serviceCfg.ObjectMeta(storage),
		}
	)

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, imgProxyDeployment, func() error {
		imgProxyDeployment.Labels = imgProxySpec.WorkloadSpec.MergeLabels(
			objectLabels(storage, serviceCfg.Name, "storage", supabase.Images.ImgProxy.Tag),
			storage.Labels,
		)

		imgProxyEnv := []corev1.EnvVar{
			serviceCfg.EnvKeys.Bind.Var(),
			serviceCfg.EnvKeys.UseETag.Var(),
			serviceCfg.EnvKeys.EnableWebPDetection.Var(imgProxySpec.EnabledWebPDetection),
		}

		if storage.Spec.Api.FileBackend != nil {
			imgProxyEnv = append(imgProxyEnv, serviceCfg.EnvKeys.LocalFileSystemRoot.Var(storage.Spec.Api.FileBackend.Path))
		}

		if imgProxyDeployment.CreationTimestamp.IsZero() {
			imgProxyDeployment.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: selectorLabels(storage, serviceCfg.Name),
			}
		}

		imgProxyDeployment.Spec.Replicas = imgProxySpec.WorkloadSpec.ReplicaCount()

		imgProxyDeployment.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: objectLabels(storage, serviceCfg.Name, "storage", supabase.Images.ImgProxy.Tag),
			},
			Spec: corev1.PodSpec{
				ImagePullSecrets:             imgProxySpec.WorkloadSpec.PullSecrets(),
				AutomountServiceAccountToken: ptrOf(false),
				Containers: []corev1.Container{{
					Name:            "supabase-imgproxy",
					Image:           imgProxySpec.WorkloadSpec.Image(supabase.Images.ImgProxy.String()),
					ImagePullPolicy: imgProxySpec.WorkloadSpec.ImagePullPolicy(),
					Env:             imgProxySpec.WorkloadSpec.MergeEnv(imgProxyEnv),
					Ports: []corev1.ContainerPort{{
						Name:          serviceCfg.Defaults.ApiPortName,
						ContainerPort: serviceCfg.Defaults.ApiPort,
						Protocol:      corev1.ProtocolTCP,
					}},
					SecurityContext: imgProxySpec.WorkloadSpec.ContainerSecurityContext(serviceCfg.Defaults.UID, serviceCfg.Defaults.GID),
					Resources:       imgProxySpec.WorkloadSpec.Resources(),
					VolumeMounts:    imgProxySpec.WorkloadSpec.AdditionalVolumeMounts(),
					ReadinessProbe: &corev1.Probe{
						InitialDelaySeconds: 5,
						PeriodSeconds:       3,
						TimeoutSeconds:      1,
						SuccessThreshold:    2,
						ProbeHandler: corev1.ProbeHandler{
							Exec: &corev1.ExecAction{
								Command: []string{"imgproxy", "health"},
							},
						},
					},
					LivenessProbe: &corev1.Probe{
						InitialDelaySeconds: 10,
						PeriodSeconds:       5,
						TimeoutSeconds:      3,
						ProbeHandler: corev1.ProbeHandler{
							Exec: &corev1.ExecAction{
								Command: []string{"imgproxy", "health"},
							},
						},
					},
				}},
				SecurityContext: imgProxySpec.WorkloadSpec.PodSecurityContext(),
				Volumes:         imgProxySpec.WorkloadSpec.Volumes(),
			},
		}

		if err := controllerutil.SetControllerReference(storage, imgProxyDeployment, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r *StorageImgProxyReconciler) reconcileImgProxyService(
	ctx context.Context,
	storage *supabasev1alpha1.Storage,
) error {
	var (
		serviceCfg      = supabase.ServiceConfig.ImgProxy
		imgProxyService = &corev1.Service{
			ObjectMeta: supabase.ServiceConfig.ImgProxy.ObjectMeta(storage),
		}
	)

	_, err := controllerutil.CreateOrPatch(ctx, r.Client, imgProxyService, func() error {
		imgProxyService.Labels = storage.Spec.Api.WorkloadSpec.MergeLabels(
			objectLabels(storage, serviceCfg.Name, "storage", supabase.Images.ImgProxy.Tag),
			storage.Labels,
		)

		imgProxyService.Spec = corev1.ServiceSpec{
			Selector: selectorLabels(storage, serviceCfg.Name),
			Ports: []corev1.ServicePort{
				{
					Name:        serviceCfg.Defaults.ApiPortName,
					Protocol:    corev1.ProtocolTCP,
					AppProtocol: ptrOf("http"),
					Port:        serviceCfg.Defaults.ApiPort,
					TargetPort:  intstr.IntOrString{IntVal: serviceCfg.Defaults.ApiPort},
				},
			},
		}

		if err := controllerutil.SetControllerReference(storage, imgProxyService, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}
