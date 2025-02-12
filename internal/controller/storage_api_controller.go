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
	"encoding/base64"
	"fmt"
	"slices"

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

// StorageApiReconciler reconciles a Storage object
type StorageApiReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.4/pkg/reconcile
func (r *StorageApiReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	if err := r.reconcileStorageApiDeployment(ctx, &storage); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileStorageApiService(ctx, &storage); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StorageApiReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&supabasev1alpha1.Storage{}).
		Named("storage-api").
		Owns(new(corev1.Secret)).
		Owns(new(appsv1.Deployment)).
		Owns(new(corev1.Service)).
		Complete(r)
}

func (r *StorageApiReconciler) reconcileStorageApiDeployment(
	ctx context.Context,
	storage *supabasev1alpha1.Storage,
) error {
	var (
		serviceCfg           = supabase.ServiceConfig.Storage
		apiSpec              = storage.Spec.Api
		storageApiDeployment = &appsv1.Deployment{
			ObjectMeta: serviceCfg.ObjectMeta(storage),
		}

		jwtSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      apiSpec.JwtAuth.SecretName,
				Namespace: storage.Namespace,
			},
		}

		s3ProtocolSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      apiSpec.S3Protocol.CredentialsSecretRef.SecretName,
				Namespace: storage.Namespace,
			},
		}

		jwtStateHash, s3ProtoCredentialsStateHash string
	)

	if err := r.Get(ctx, client.ObjectKeyFromObject(jwtSecret), jwtSecret); err != nil {
		return err
	}

	jwtStateHash = base64.StdEncoding.EncodeToString(HashBytes(
		jwtSecret.Data[apiSpec.JwtAuth.SecretKey],
		jwtSecret.Data[apiSpec.JwtAuth.JwksKey],
	))

	if err := r.Get(ctx, client.ObjectKeyFromObject(s3ProtocolSecret), s3ProtocolSecret); err != nil {
		return err
	}

	s3ProtoCredentialsStateHash = base64.StdEncoding.EncodeToString(HashBytes(
		s3ProtocolSecret.Data[apiSpec.S3Protocol.CredentialsSecretRef.AccessKeyIdKey],
		s3ProtocolSecret.Data[apiSpec.S3Protocol.CredentialsSecretRef.AccessSecretKeyKey],
	))

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, storageApiDeployment, func() error {
		storageApiDeployment.Labels = apiSpec.WorkloadSpec.MergeLabels(
			objectLabels(storage, serviceCfg.Name, "storage", supabase.Images.Storage.Tag),
			storage.Labels,
		)

		storagApiEnv := []corev1.EnvVar{
			{
				Name: "DB_USERNAME",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: apiSpec.DBSpec.DBCredentialsRef.SecretName,
						},
						Key: apiSpec.DBSpec.DBCredentialsRef.UsernameKey,
					},
				},
			},
			{
				Name: "DB_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: apiSpec.DBSpec.DBCredentialsRef.SecretName,
						},
						Key: apiSpec.DBSpec.DBCredentialsRef.PasswordKey,
					},
				},
			},
			serviceCfg.EnvKeys.DatabaseDSN.Var(fmt.Sprintf("postgres://$(DB_USERNAME):$(DB_PASSWORD)@%s:%d/%s", apiSpec.DBSpec.Host, apiSpec.DBSpec.Port, apiSpec.DBSpec.DBName)),
			serviceCfg.EnvKeys.ServiceKey.Var(apiSpec.JwtAuth.ServiceKeySelector()),
			serviceCfg.EnvKeys.JwtSecret.Var(apiSpec.JwtAuth.SecretKeySelector()),
			serviceCfg.EnvKeys.JwtJwks.Var(apiSpec.JwtAuth.JwksKeySelector()),
			serviceCfg.EnvKeys.S3ProtocolPrefix.Var(),
			serviceCfg.EnvKeys.S3ProtocolAllowForwardedHeader.Var(apiSpec.S3Protocol.AllowForwardedHeader),
			serviceCfg.EnvKeys.S3ProtocolAccessKeyId.Var(apiSpec.S3Protocol.CredentialsSecretRef.AccessKeyIdSelector()),
			serviceCfg.EnvKeys.S3ProtocolAccessKeySecret.Var(apiSpec.S3Protocol.CredentialsSecretRef.AccessSecretKeySelector()),
			serviceCfg.EnvKeys.TusUrlPath.Var(),
			serviceCfg.EnvKeys.FileSizeLimit.Var(apiSpec.FileSizeLimit),
			serviceCfg.EnvKeys.UploadFileSizeLimit.Var(apiSpec.FileSizeLimit),
			serviceCfg.EnvKeys.UploadFileSizeLimitStandard.Var(apiSpec.FileSizeLimit),
			serviceCfg.EnvKeys.AnonKey.Var(apiSpec.JwtAuth.AnonKeySelector()),
			// TODO: https://github.com/supabase/storage-api/issues/55
			serviceCfg.EnvKeys.FileStorageRegion.Var(),
		}

		if storage.Spec.ImageProxy != nil && storage.Spec.ImageProxy.Enable {
			storagApiEnv = append(storagApiEnv, serviceCfg.EnvKeys.ImgProxyURL.Var(fmt.Sprintf("http://%s.%s.svc:%d", supabase.ServiceConfig.ImgProxy.ObjectName(storage), storage.Namespace, supabase.ServiceConfig.ImgProxy.Defaults.ApiPort)))
		}

		if storageApiDeployment.CreationTimestamp.IsZero() {
			storageApiDeployment.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: selectorLabels(storage, serviceCfg.Name),
			}
		}

		storageApiDeployment.Spec.Replicas = apiSpec.WorkloadSpec.ReplicaCount()

		storageApiDeployment.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					fmt.Sprintf("%s/%s", supabasev1alpha1.GroupVersion.Group, "jwt-hash"):            jwtStateHash,
					fmt.Sprintf("%s/%s", supabasev1alpha1.GroupVersion.Group, "s3-credentials-hash"): s3ProtoCredentialsStateHash,
				},
				Labels: objectLabels(storage, serviceCfg.Name, "storage", supabase.Images.Storage.Tag),
			},
			Spec: corev1.PodSpec{
				ImagePullSecrets: apiSpec.WorkloadSpec.PullSecrets(),
				Containers: []corev1.Container{{
					Name:            "supabase-storage",
					Image:           apiSpec.WorkloadSpec.Image(supabase.Images.Storage.String()),
					ImagePullPolicy: apiSpec.WorkloadSpec.ImagePullPolicy(),
					Env:             apiSpec.WorkloadSpec.MergeEnv(append(storagApiEnv, slices.Concat(apiSpec.FileBackend.Env(), apiSpec.S3Backend.Env())...)),
					Ports: []corev1.ContainerPort{{
						Name:          serviceCfg.Defaults.ApiPortName,
						ContainerPort: serviceCfg.Defaults.ApiPort,
						Protocol:      corev1.ProtocolTCP,
					}},
					SecurityContext: apiSpec.WorkloadSpec.ContainerSecurityContext(serviceCfg.Defaults.UID, serviceCfg.Defaults.GID),
					Resources:       apiSpec.WorkloadSpec.Resources(),
					VolumeMounts: apiSpec.WorkloadSpec.AdditionalVolumeMounts(
						corev1.VolumeMount{
							Name:      "tmp",
							MountPath: "/tmp",
						},
					),
					ReadinessProbe: &corev1.Probe{
						InitialDelaySeconds: 5,
						PeriodSeconds:       3,
						TimeoutSeconds:      1,
						SuccessThreshold:    2,
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: serviceCfg.LivenessProbePath,
								Port: intstr.IntOrString{IntVal: serviceCfg.Defaults.ApiPort},
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
								Port: intstr.IntOrString{IntVal: serviceCfg.Defaults.ApiPort},
							},
						},
					},
				}},
				SecurityContext: apiSpec.WorkloadSpec.PodSecurityContext(),
				Volumes: apiSpec.WorkloadSpec.Volumes(
					corev1.Volume{
						Name: "tmp",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: apiSpec.UploadTemp.VolumeSource(),
						},
					},
				),
			},
		}

		if err := controllerutil.SetControllerReference(storage, storageApiDeployment, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r *StorageApiReconciler) reconcileStorageApiService(
	ctx context.Context,
	storage *supabasev1alpha1.Storage,
) error {
	var (
		serviceCfg        = supabase.ServiceConfig.Storage
		storageApiService = &corev1.Service{
			ObjectMeta: supabase.ServiceConfig.Storage.ObjectMeta(storage),
		}
	)

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, storageApiService, func() error {
		storageApiService.Labels = storage.Spec.Api.WorkloadSpec.MergeLabels(
			objectLabels(storage, serviceCfg.Name, "storage", supabase.Images.Storage.Tag),
			storage.Labels,
		)

		if _, ok := storageApiService.Labels[meta.SupabaseLabel.ApiGatewayTarget]; !ok {
			storageApiService.Labels[meta.SupabaseLabel.ApiGatewayTarget] = ""
		}

		storageApiService.Spec = corev1.ServiceSpec{
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

		if err := controllerutil.SetControllerReference(storage, storageApiService, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}
