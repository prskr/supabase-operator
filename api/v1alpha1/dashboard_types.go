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

type StudioSpec struct {
	JWT *JwtSpec `json:"jwt,omitempty"`
	// WorkloadTemplate - customize the studio deployment
	WorkloadTemplate *WorkloadTemplate `json:"workloadTemplate,omitempty"`
	// GatewayServiceSelector - selector to find the service for the API gateway
	// Required to configure the API URL in the studio deployment
	// If you don't run multiple APIGateway instances in the same namespaces, the default will be fine
	// +kubebuilder:default={"app.kubernetes.io/name":"envoy","app.kubernetes.io/component":"api-gateway"}
	GatewayServiceMatchLabels map[string]string `json:"gatewayServiceSelector,omitempty"`
	// APIExternalURL is referring to the URL where Supabase API will be available
	// Typically this is the ingress of the API gateway
	APIExternalURL string `json:"externalUrl"`
}

type PGMetaSpec struct {
	// WorkloadTemplate - customize the pg-meta deployment
	WorkloadTemplate *WorkloadTemplate `json:"workloadTemplate,omitempty"`
}

type DbCredentialsReference struct {
	SecretName string `json:"secretName"`
	// UsernameKey
	// +kubebuilder:default="username"
	UsernameKey string `json:"usernameKey,omitempty"`
	// PasswordKey
	// +kubebuilder:default="password"
	PasswordKey string `json:"passwordKey,omitempty"`
}

type DashboardDbSpec struct {
	Host string `json:"host"`
	// Port - Database port, typically 5432
	// +kubebuilder:default=5432
	Port   int    `json:"port,omitempty"`
	DBName string `json:"dbName"`
	// DBCredentialsRef - reference to a Secret key where the DB credentials can be retrieved from
	// Credentials need to be stored in basic auth form
	DBCredentialsRef *DbCredentialsReference `json:"dbCredentialsRef"`
}

func (s DashboardDbSpec) UserRef() *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: s.DBCredentialsRef.SecretName,
		},
		Key: s.DBCredentialsRef.UsernameKey,
	}
}

func (s DashboardDbSpec) PasswordRef() *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: s.DBCredentialsRef.SecretName,
		},
		Key: s.DBCredentialsRef.PasswordKey,
	}
}

// DashboardSpec defines the desired state of Dashboard.
type DashboardSpec struct {
	DBSpec *DashboardDbSpec `json:"db"`
	// PGMeta
	PGMeta *PGMetaSpec `json:"pgMeta,omitempty"`
	// Studio
	Studio *StudioSpec `json:"studio,omitempty"`
}

// DashboardStatus defines the observed state of Dashboard.
type DashboardStatus struct{}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Dashboard is the Schema for the dashboards API.
type Dashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DashboardSpec   `json:"spec,omitempty"`
	Status DashboardStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DashboardList contains a list of Dashboard.
type DashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dashboard `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Dashboard{}, &DashboardList{})
}
