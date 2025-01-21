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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StorageBackend string

const (
	StorageBackendFile StorageBackend = "file"
	StorageBackendS3   StorageBackend = "s3"
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

type S3ProtocolSpec struct {
	// Region - S3 region to use in the API
	// +kubebuilder:default="us-east-1"
	Region string `json:"region,omitempty"`

	// AllowForwardedHeader
	// +kubebuilder:default=true
	AllowForwardedHeader bool `json:"allowForwardedHeader,omitempty"`

	// CredentialsSecretRef - reference to the Secret where access key id and access secret key are stored
	CredentialsSecretRef *S3CredentialsRef `json:"credentialsSecretRef,omitempty"`
}

// StorageSpec defines the desired state of Storage.
type StorageSpec struct {
	// BackendType - backend storage type to use
	// +kubebuilder:validation:Enum={s3,file}
	BackendType StorageBackend `json:"backendType"`
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
	// S3 - Configure S3 protocol
	S3 *S3ProtocolSpec `json:"s3,omitempty"`
	// EnableImageTransformation - whether to deploy the image proxy
	// the image proxy scale images to lower resolutions on demand to reduce traffic for instance for mobile devices
	EnableImageTransformation bool `json:"enableImageTransformation,omitempty"`
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
