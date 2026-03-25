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

func init() {
	SchemeBuilder.Register(&Dashboard{}, &DashboardList{})
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Dashboard is the Schema for the dashboards API.
type Dashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DashboardSpec   `json:"spec,omitempty"`
	Status DashboardStatus `json:"status,omitempty"`
}

// DashboardStatus defines the observed state of Dashboard.
type DashboardStatus struct{}

// DashboardSpec defines the desired state of Dashboard.
type DashboardSpec struct {
	DBSpec *DashboardDBSpec `json:"db"`
	// PGMeta
	PGMeta *PGMetaSpec `json:"pgMeta,omitempty"`
	// Studio
	Studio *StudioSpec `json:"studio,omitempty"`
}

type DashboardDBSpec struct {
	Host string `json:"host"`
	// Port - Database port, typically 5432
	// +kubebuilder:default=5432
	Port   int    `json:"port,omitempty"`
	DBName string `json:"dbName"`
	// DBCredentialsRef - reference to a Secret key where the DB credentials can be retrieved from
	// Credentials need to be stored in basic auth form
	DBCredentialsRef *DBCredentialsReference `json:"dbCredentialsRef"`
	// Schemas - schema where PostgREST is looking for objects (tables, views, functions, ...)
	// +kubebuilder:default={"public","storage","graphql_public"}
	Schemas []string `json:"schemas,omitempty"`
	// ExtraSearchPath - Extra schemas to add to the search_path of every request.
	// These schemas tables, views and functions don’t get API endpoints, they can only be referred from the database objects inside your db-schemas.
	// +kubebuilder:default={"public"}
	ExtraSearchPath []string `json:"extraSearchPath,omitempty"`
}

func (s DashboardDBSpec) UserRef() *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: s.DBCredentialsRef.SecretName,
		},
		Key: s.DBCredentialsRef.UsernameKey,
	}
}

func (s DashboardDBSpec) PasswordRef() *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: s.DBCredentialsRef.SecretName,
		},
		Key: s.DBCredentialsRef.PasswordKey,
	}
}

type DBCredentialsReference struct {
	SecretName string `json:"secretName"`
	// UsernameKey
	// +kubebuilder:default="username"
	UsernameKey string `json:"usernameKey,omitempty"`
	// PasswordKey
	// +kubebuilder:default="password"
	PasswordKey string `json:"passwordKey,omitempty"`
}

type PGMetaSpec struct {
	// WorkloadTemplate - customize the pg-meta deployment
	WorkloadSpec *WorkloadSpec `json:"workloadSpec,omitempty"`
}

type StudioSpec struct {
	// AI / LLM integration
	// configure OpenAI API key and optionally base URL
	AI  *AISpec  `json:"ai,omitempty"`
	JWT *JwtSpec `json:"jwt,omitempty"`
	// WorkloadTemplate - customize the studio deployment
	WorkloadSpec *WorkloadSpec `json:"workloadSpec,omitempty"`
	// GatewayServiceSelector - selector to find the service for the API gateway
	// Required to configure the API URL in the studio deployment
	// If you don't run multiple APIGateway instances in the same namespaces, the default will be fine
	// +kubebuilder:default={"app.kubernetes.io/name":"envoy","app.kubernetes.io/component":"api-gateway"}
	GatewayServiceMatchLabels map[string]string `json:"gatewayServiceSelector,omitempty"`
	// APIExternalURL is referring to the URL where Supabase API will be available
	// Typically this is the ingress of the API gateway
	APIExternalURL string `json:"externalUrl"`
}

type AISpec struct {
	OpenAI *OpenAISpec `json:"openai,omitempty"`
}

type OpenAISpec struct {
	APIKey *corev1.SecretKeySelector `json:"apiKey"`
}

func (s *OpenAISpec) Vars() []corev1.EnvVar {
	if s == nil {
		return nil
	}

	vars := []corev1.EnvVar{
		{
			Name: "OPENAI_API_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: s.APIKey,
			},
		},
	}

	return vars
}

// +kubebuilder:object:root=true

// DashboardList contains a list of Dashboard.
type DashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dashboard `json:"items"`
}
