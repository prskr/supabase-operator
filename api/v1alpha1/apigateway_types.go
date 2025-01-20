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
	"iter"
	"maps"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&APIGateway{}, &APIGatewayList{})
}

type ControlPlaneSpec struct {
	// Host is the hostname of the envoy control plane endpoint
	Host string `json:"host"`
	// Port is the port number of the envoy control plane endpoint - typically this is 18000
	// +kubebuilder:default=18000
	// +kubebuilder:validation:Maximum=65535
	Port uint16 `json:"port"`
}

type EnvoySpec struct {
	// NodeName - identifies the Envoy cluster within the current namespace
	// if not set, the name of the APIGateway resource will be used
	// The primary use case is to make the assignment of multiple supabase instances in a single namespace explicit.
	NodeName string `json:"nodeName,omitempty"`
	// ControlPlane - configure the control plane where Envoy will retrieve its configuration from
	ControlPlane *ControlPlaneSpec `json:"controlPlane"`
	// WorkloadTemplate - customize the Envoy deployment
	WorkloadTemplate *WorkloadTemplate `json:"workloadTemplate,omitempty"`
}

// APIGatewaySpec defines the desired state of APIGateway.
type APIGatewaySpec struct {
	// Envoy - configure the envoy instance and most importantly the control-plane
	Envoy *EnvoySpec `json:"envoy"`
	// JWKSSelector - selector where the JWKS can be retrieved from to enable the API gateway to validate JWTs
	JWKSSelector *corev1.SecretKeySelector `json:"jwks"`
	// ServiceSelector - selector to match all Supabase services (or in fact EndpointSlices) that should be considered for this APIGateway
	// +kubebuilder:default={"matchExpressions":{{"key": "app.kubernetes.io/part-of", "operator":"In", "values":{"supabase"}},{"key":"supabase.k8s.icb4dc0.de/api-gateway-target","operator":"Exists"}}}
	ServiceSelector *metav1.LabelSelector `json:"serviceSelector"`
	// ComponentTypeLabel - Label to identify which Supabase component a Service represents (e.g. auth, postgrest, ...)
	// +kubebuilder:default="app.kubernetes.io/name"
	ComponentTypeLabel string `json:"componentTypeLabel,omitempty"`
}

type EnvoyStatus struct {
	ConfigVersion string `json:"configVersion,omitempty"`
	ResourceHash  []byte `json:"resourceHash,omitempty"`
}

// APIGatewayStatus defines the observed state of APIGateway.
type APIGatewayStatus struct {
	Envoy          EnvoyStatus         `json:"envoy,omitempty"`
	ServiceTargets map[string][]string `json:"serviceTargets,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// APIGateway is the Schema for the apigateways API.
// +kubebuilder:printcolumn:name="EnvoyConfigVersion",type=string,JSONPath=`.status.envoy.configVersion`
type APIGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIGatewaySpec   `json:"spec,omitempty"`
	Status APIGatewayStatus `json:"status,omitempty"`
}

func (g APIGateway) JwksSecretMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      g.Spec.JWKSSelector.Name,
		Namespace: g.Namespace,
		Labels:    maps.Clone(g.Labels),
	}
}

// +kubebuilder:object:root=true

// APIGatewayList contains a list of APIGateway.
type APIGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIGateway `json:"items"`
}

func (l APIGatewayList) Iter() iter.Seq[*APIGateway] {
	return func(yield func(*APIGateway) bool) {
		for _, gw := range l.Items {
			if !yield(&gw) {
				return
			}
		}
	}
}
