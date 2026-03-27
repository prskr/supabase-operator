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
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	supabasev1alpha1 "github.com/prskr/supabase-operator/api/v1alpha1"
	"github.com/prskr/supabase-operator/internal/meta"
	"github.com/prskr/supabase-operator/internal/supabase"
)

// StorageAPIReconciler reconciles a Storage object
type StorageAPIReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder events.EventRecorder
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.4/pkg/reconcile
func (r *StorageAPIReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	if err := r.reconcileStorageAPIDeployment(ctx, &storage); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileStorageAPIService(ctx, &storage); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StorageAPIReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&supabasev1alpha1.Storage{}).
		Named("storage-api").
		Owns(new(corev1.Secret)).
		Owns(new(appsv1.Deployment)).
		Owns(new(corev1.Service)).
		Complete(r)
}

func (r *StorageAPIReconciler) reconcileStorageAPIDeployment(
	ctx context.Context,
	storage *supabasev1alpha1.Storage,
) error {
	var (
		serviceCfg           = supabase.ServiceConfig.Storage
		apiSpec              = storage.Spec.API
		storageAPIDeployment = &appsv1.Deployment{
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
		s3ProtocolSecret.Data[apiSpec.S3Protocol.CredentialsSecretRef.AccessKeyIDKey],
		s3ProtocolSecret.Data[apiSpec.S3Protocol.CredentialsSecretRef.AccessSecretKeyKey],
	))

	var postgrestURL string
	if len(apiSpec.PostgRESTServiceMatchLabels) > 0 {
		var serviceList corev1.ServiceList
		if err := r.List(ctx, &serviceList, client.InNamespace(storage.Namespace), client.MatchingLabels(apiSpec.PostgRESTServiceMatchLabels)); err != nil {
			return err
		}

		if len(serviceList.Items) == 1 {
			svc := serviceList.Items[0]
			port := supabase.ServiceConfig.Postgrest.Defaults.ServerPort
			for _, p := range svc.Spec.Ports {
				if p.Name == supabase.ServiceConfig.Postgrest.Defaults.ServerPortName {
					port = p.Port
					break
				}
			}
			postgrestURL = fmt.Sprintf("http://%s.%s.svc:%d", svc.Name, svc.Namespace, port)
		} else {
			r.Recorder.Eventf(
				storage,
				storage,
				corev1.EventTypeWarning,
				"AmbiguousPostgRESTService",
				"The link to the PostgREST service cannot be determined",
				"Make sure that the configured selector matches a single service",
			)
		}
	} else {
		r.Recorder.Eventf(
			storage,
			storage,
			corev1.EventTypeWarning,
			"PostgRESTServiceSelectorEmpty",
			"PostgRESTServiceMatchLabels is empty, PostgREST URL will not be configured",
			"This should never happen because the selector has a sane default",
		)
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, storageAPIDeployment, func() error {
		storageAPIDeployment.Labels = apiSpec.WorkloadSpec.MergeLabels(
			objectLabels(storage, serviceCfg.Name, "storage", supabase.Images.Storage.Tag),
			storage.Labels,
		)

		storagAPIEnv := []corev1.EnvVar{
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
			serviceCfg.EnvKeys.PostgrestURL.Var(postgrestURL),
			serviceCfg.EnvKeys.S3ProtocolPrefix.Var(),
			serviceCfg.EnvKeys.S3ProtocolAllowForwardedHeader.Var(apiSpec.S3Protocol.AllowForwardedHeader),
			serviceCfg.EnvKeys.S3ProtocolAccessKeyID.Var(apiSpec.S3Protocol.CredentialsSecretRef.AccessKeyIDSelector()),
			serviceCfg.EnvKeys.S3ProtocolAccessKeySecret.Var(apiSpec.S3Protocol.CredentialsSecretRef.AccessSecretKeySelector()),
			serviceCfg.EnvKeys.FileSizeLimit.Var(apiSpec.FileSizeLimit),
			serviceCfg.EnvKeys.UploadFileSizeLimit.Var(apiSpec.FileSizeLimit),
			serviceCfg.EnvKeys.UploadFileSizeLimitStandard.Var(apiSpec.FileSizeLimit),
			serviceCfg.EnvKeys.AnonKey.Var(apiSpec.JwtAuth.AnonKeySelector()),
			// TODO: https://github.com/supabase/storage-api/issues/55
			serviceCfg.EnvKeys.FileStorageRegion.Var(*apiSpec.Region),
			serviceCfg.EnvKeys.TenantID.Var(*apiSpec.TenantID),
			serviceCfg.EnvKeys.AllowForwardedPathHeader.Var(),
			serviceCfg.EnvKeys.PGQeueEnabled.Var(false),
		}

		if storage.Spec.ImageProxy != nil && storage.Spec.ImageProxy.Enable {
			storagAPIEnv = append(storagAPIEnv, serviceCfg.EnvKeys.ImgProxyURL.Var(fmt.Sprintf("http://%s.%s.svc:%d", supabase.ServiceConfig.ImgProxy.ObjectName(storage), storage.Namespace, supabase.ServiceConfig.ImgProxy.Defaults.ApiPort)))
		}

		if storageAPIDeployment.CreationTimestamp.IsZero() {
			storageAPIDeployment.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: selectorLabels(storage, serviceCfg.Name),
			}
		}

		storageAPIDeployment.Spec.Replicas = apiSpec.WorkloadSpec.ReplicaCount()

		storageAPIDeployment.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					fmt.Sprintf("%s/%s", supabasev1alpha1.GroupVersion.Group, "jwt-hash"):            jwtStateHash,
					fmt.Sprintf("%s/%s", supabasev1alpha1.GroupVersion.Group, "s3-credentials-hash"): s3ProtoCredentialsStateHash,
				},
				Labels: objectLabels(storage, serviceCfg.Name, "storage", supabase.Images.Storage.Tag),
			},
			Spec: corev1.PodSpec{
				ImagePullSecrets:             apiSpec.WorkloadSpec.PullSecrets(),
				AutomountServiceAccountToken: new(false),
				Containers: []corev1.Container{{
					Name:            "supabase-storage",
					Image:           apiSpec.WorkloadSpec.Image(supabase.Images.Storage.String()),
					ImagePullPolicy: apiSpec.WorkloadSpec.ImagePullPolicy(),
					Env:             apiSpec.WorkloadSpec.MergeEnv(append(storagAPIEnv, slices.Concat(apiSpec.FileBackend.Env(), apiSpec.S3Backend.Env())...)),
					Ports: []corev1.ContainerPort{{
						Name:          serviceCfg.Defaults.APIPortName,
						ContainerPort: serviceCfg.Defaults.APIPort,
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

		if err := controllerutil.SetControllerReference(storage, storageAPIDeployment, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r *StorageAPIReconciler) reconcileStorageAPIService(
	ctx context.Context,
	storage *supabasev1alpha1.Storage,
) error {
	var (
		serviceCfg        = supabase.ServiceConfig.Storage
		storageAPIService = &corev1.Service{
			ObjectMeta: supabase.ServiceConfig.Storage.ObjectMeta(storage),
		}
	)

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, storageAPIService, func() error {
		storageAPIService.Labels = storage.Spec.API.WorkloadSpec.MergeLabels(
			objectLabels(storage, serviceCfg.Name, "storage", supabase.Images.Storage.Tag),
			storage.Labels,
		)

		if _, ok := storageAPIService.Labels[meta.SupabaseLabel.ApiGatewayTarget]; !ok {
			storageAPIService.Labels[meta.SupabaseLabel.ApiGatewayTarget] = ""
		}

		storageAPIService.Spec = corev1.ServiceSpec{
			Selector: selectorLabels(storage, serviceCfg.Name),
			Ports: []corev1.ServicePort{
				{
					Name:        serviceCfg.Defaults.APIPortName,
					Protocol:    corev1.ProtocolTCP,
					AppProtocol: new("http"),
					Port:        serviceCfg.Defaults.APIPort,
					TargetPort:  intstr.IntOrString{IntVal: serviceCfg.Defaults.APIPort},
				},
			},
		}

		if err := controllerutil.SetControllerReference(storage, storageAPIService, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}
