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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"code.icb4dc0.de/prskr/supabase-operator/internal/supabase"
)

type BackendStorageType string

const (
	BackendStorageTypeFile BackendStorageType = "file"
	BackendStorageTypeS3   BackendStorageType = "s3"
)

type StorageApiDbSpec struct {
	Host string `json:"host"`
	// Port - Database port, typically 5432
	// +kubebuilder:default=5432
	Port   int    `json:"port,omitempty"`
	DBName string `json:"dbName"`
	// DBCredentialsRef - reference to a Secret key where the DB credentials can be retrieved from
	// Credentials need to be stored in basic auth form
	DBCredentialsRef *DbCredentialsReference `json:"dbCredentialsRef"`
}

func (s StorageApiDbSpec) UserRef() *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: s.DBCredentialsRef.SecretName,
		},
		Key: s.DBCredentialsRef.UsernameKey,
	}
}

func (s StorageApiDbSpec) PasswordRef() *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: s.DBCredentialsRef.SecretName,
		},
		Key: s.DBCredentialsRef.PasswordKey,
	}
}

type S3CredentialsRef struct {
	SecretName string `json:"secretName"`
	// AccessKeyIdKey - key in Secret where access key id will be referenced from
	// +kubebuilder:default="accessKeyId"
	AccessKeyIdKey string `json:"accessKeyIdKey,omitempty"`
	// AccessSecretKeyKey - key in Secret where access secret key will be referenced from
	// +kubebuilder:default="secretAccessKey"
	AccessSecretKeyKey string `json:"accessSecretKeyKey,omitempty"`
}

func (r S3CredentialsRef) AccessKeyIdSelector() *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: r.SecretName,
		},
		Key: r.AccessKeyIdKey,
	}
}

func (r S3CredentialsRef) AccessSecretKeySelector() *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: r.SecretName,
		},
		Key: r.AccessSecretKeyKey,
	}
}

type S3ProtocolSpec struct {
	// AllowForwardedHeader
	// +kubebuilder:default=true
	AllowForwardedHeader bool `json:"allowForwardedHeader,omitempty"`

	// CredentialsSecretRef - reference to the Secret where access key id and access secret key are stored
	CredentialsSecretRef *S3CredentialsRef `json:"credentialsSecretRef,omitempty"`
}

type FileBackendSpec struct {
	// Path - path to where files will be stored
	Path string `json:"path"`
}

func (s *FileBackendSpec) Env() []corev1.EnvVar {
	if s == nil {
		return nil
	}

	svcCfg := supabase.ServiceConfig.Storage

	return []corev1.EnvVar{
		svcCfg.EnvKeys.StorageBackend.Var("file"),
		svcCfg.EnvKeys.TenantID.Var(),
		svcCfg.EnvKeys.FileStorageBackendPath.Var(s.Path),
		svcCfg.EnvKeys.StorageS3Region.Var("local"),
		svcCfg.EnvKeys.StorageS3Bucket.Var("stub"),
	}
}

type S3BackendSpec struct {
	// Region - S3 region of the backend
	Region string `json:"region"`
	// Endpoint - hostname and port **with** http/https
	Endpoint string `json:"endpoint"`
	// ForcePathStyle - whether to use path style (e.g. for MinIO) or domain style
	// for bucket addressing
	ForcePathStyle bool `json:"forcePathStyle,omitempty"`
	// Bucket - bucke to use, if file backend is used, default value is sufficient
	// +kubebuilder:default="stub"
	Bucket string `json:"bucket"`

	// CredentialsSecretRef - reference to the Secret where access key id and access secret key are stored
	CredentialsSecretRef *S3CredentialsRef `json:"credentialsSecretRef"`
}

func (s *S3BackendSpec) Env() []corev1.EnvVar {
	if s == nil {
		return nil
	}

	svcCfg := supabase.ServiceConfig.Storage

	return []corev1.EnvVar{
		svcCfg.EnvKeys.StorageBackend.Var("s3"),
		svcCfg.EnvKeys.StorageS3Endpoint.Var(s.Endpoint),
		svcCfg.EnvKeys.StorageS3ForcePathStyle.Var(s.ForcePathStyle),
		svcCfg.EnvKeys.StorageS3Bucket.Var(s.Bucket),
		svcCfg.EnvKeys.StorageS3Region.Var(s.Region),
		svcCfg.EnvKeys.StorageS3AccessKeyId.Var(s.CredentialsSecretRef.AccessKeyIdSelector()),
		svcCfg.EnvKeys.StorageS3AccessSecretKey.Var(s.CredentialsSecretRef.AccessSecretKeySelector()),
	}
}

type UploadTempSpec struct {
	// Medium of the empty dir to cache uploads
	Medium    corev1.StorageMedium `json:"medium,omitempty"`
	SizeLimit *resource.Quantity   `json:"sizeLimit,omitempty"`
}

func (s *UploadTempSpec) VolumeSource() *corev1.EmptyDirVolumeSource {
	if s == nil {
		return &corev1.EmptyDirVolumeSource{
			Medium: corev1.StorageMediumDefault,
		}
	}

	return &corev1.EmptyDirVolumeSource{
		Medium:    s.Medium,
		SizeLimit: s.SizeLimit,
	}
}

type StorageApiSpec struct {
	S3Backend *S3BackendSpec `json:"s3Backend,omitempty"`
	// FileBackend - configure the file backend
	// either S3 or file backend **MUST** be configured
	FileBackend *FileBackendSpec `json:"fileBackend,omitempty"`
	// FileSizeLimit - maximum file upload size in bytes
	// +kubebuilder:default=52428800
	FileSizeLimit uint64 `json:"fileSizeLimit,omitempty"`
	// JwtAuth - Configure the JWT authentication parameters.
	// This includes where to retrieve anon and service key from as well as JWT secret and JWKS references
	// needed to validate JWTs send to the API
	JwtAuth JwtSpec `json:"jwtAuth"`
	// DBSpec - Configure access to the Postgres database
	// In most cases this will reference the supabase-storage-admin credentials secret provided by the Core resource
	DBSpec StorageApiDbSpec `json:"db"`
	// S3Protocol - Configure S3 access to the Storage API allowing clients to use any S3 client
	S3Protocol *S3ProtocolSpec `json:"s3,omitempty"`
	// UploadTemp - configure the emptyDir for storing intermediate files during uploads
	UploadTemp *UploadTempSpec `json:"uploadTemp,omitempty"`
	// WorkloadTemplate - customize the Storage API workload
	WorkloadSpec *WorkloadSpec `json:"workloadSpec,omitempty"`
}

type ImageProxySpec struct {
	// Enable - whether to deploy the image proxy or not
	Enable               bool `json:"enable,omitempty"`
	EnabledWebPDetection bool `json:"enableWebPDetection,omitempty"`
	// WorkloadTemplate - customize the image proxy workload
	WorkloadSpec *WorkloadSpec `json:"workloadSpec,omitempty"`
}

// StorageSpec defines the desired state of Storage.
type StorageSpec struct {
	// Api - configure the Storage API
	Api StorageApiSpec `json:"api,omitempty"`

	// ImageProxy - optionally enable and configure the image proxy
	// the image proxy scale images to lower resolutions on demand to reduce traffic for instance for mobile devices
	ImageProxy *ImageProxySpec `json:"imageProxy,omitempty"`
}

// StorageStatus defines the observed state of Storage.
type StorageStatus struct{}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Storage is the Schema for the storages API.
type Storage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StorageSpec   `json:"spec,omitempty"`
	Status StorageStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StorageList contains a list of Storage.
type StorageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Storage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Storage{}, &StorageList{})
}
