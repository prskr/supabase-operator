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
	"strings"

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

type EnvoyLogLevel string

type EnvoyComponentLogLevel struct {
	// Component - the component to set the log level for
	// the component IDs can be found [here](https://github.com/envoyproxy/envoy/blob/main/source/common/common/logger.h#L36)
	Component string `json:"component"`
	// Level - the log level to set for the component
	// +kubebuilder:validation:Enum=trace;debug;info;warning;error;critical;off
	Level EnvoyLogLevel `json:"level"`
}

type EnvoyDebuggingOptions struct {
	ComponentLogLevels []EnvoyComponentLogLevel `json:"componentLogLevels,omitempty"`
}

func (o *EnvoyDebuggingOptions) DebugLogging() string {
	if o == nil || len(o.ComponentLogLevels) == 0 {
		return ""
	}

	var builder strings.Builder
	for i, lvl := range o.ComponentLogLevels {
		if i > 0 {
			builder.WriteString(",")
		}

		builder.WriteString(lvl.Component)
		builder.WriteRune(':')
		builder.WriteString(string(lvl.Level))
	}

	return builder.String()
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
	// DisableIPv6 - disable IPv6 for the Envoy instance
	// this will force Envoy to use IPv4 for upstream hosts (mostly for the OAuth2 token endpoint)
	DisableIPv6 bool                   `json:"disableIPv6,omitempty"`
	Debugging   *EnvoyDebuggingOptions `json:"debugging,omitempty"`
}

type TlsCertRef struct {
	SecretName string `json:"secretName"`
	// ServerCertKey - key in the secret that contains the server certificate
	// +kubebuilder:default="tls.crt"
	ServerCertKey string `json:"serverCertKey"`
	// ServerKeyKey - key in the secret that contains the server private key
	// +kubebuilder:default="tls.key"
	ServerKeyKey string `json:"serverKeyKey"`
	// CaCertKey - key in the secret that contains the CA certificate
	// +kubebuilder:default="ca.crt"
	CaCertKey string `json:"caCertKey,omitempty"`
}

type EndpointTlsSpec struct {
	Cert *TlsCertRef `json:"cert"`
}

type ApiEndpointSpec struct {
	// JWKSSelector - selector where the JWKS can be retrieved from to enable the API gateway to validate JWTs
	JWKSSelector *corev1.SecretKeySelector `json:"jwks"`
	// TLS - enable and configure TLS for the API endpoint
	TLS *EndpointTlsSpec `json:"tls,omitempty"`
}

func (s *ApiEndpointSpec) TLSSpec() *EndpointTlsSpec {
	if s == nil {
		return nil
	}

	return s.TLS
}

type DashboardAuthType string

const (
	DashboardAuthTypeNone   DashboardAuthType = "none"
	DashboardAuthTypeOAuth2 DashboardAuthType = "oauth2"
	DashboardAuthTypeBasic  DashboardAuthType = "basic"
)

type DashboardOAuth2Spec struct {
	// OpenIDIssuer - if set the defaulter will fetch the discovery document and fill
	// TokenEndpoint and AuthorizationEndpoint based on the discovery document
	OpenIDIssuer string `json:"openIdIssuer,omitempty"`
	// TokenEndpoint - endpoint where Envoy will retrieve the OAuth2 access and identity token from
	TokenEndpoint string `json:"tokenEndpoint,omitempty"`
	// AuthorizationEndpoint - endpoint where the user will be redirected to authenticate
	AuthorizationEndpoint string `json:"authorizationEndpoint,omitempty"`
	// ClientID - client ID to authenticate with the OAuth2 provider
	ClientID string `json:"clientId"`
	// Scopes - scopes to request from the OAuth2 provider (e.g. "openid", "profile", ...) - optional
	Scopes []string `json:"scopes,omitempty"`
	// Resources - resources to request from the OAuth2 provider (e.g. "user", "email", ...) - optional
	Resources []string `json:"resources,omitempty"`
	// ClientSecretRef - reference to the secret that contains the client secret
	ClientSecretRef *corev1.SecretKeySelector `json:"clientSecretRef"`
}

type DashboardBasicAuthSpec struct {
	// UsersInline - [htpasswd format](https://httpd.apache.org/docs/2.4/programs/htpasswd.html)
	// +kubebuilder:validation:items:Pattern="^[\\w_.]+:\\{SHA\\}[A-z0-9]+=*$"
	UsersInline []string `json:"usersInline,omitempty"`
	// PlaintextUsersSecretRef - name of a secret that contains plaintext credentials in key-value form
	// if not empty, credentials will be merged with inline users
	PlaintextUsersSecretRef string `json:"plaintextUsersSecretRef,omitempty"`
}

type DashboardAuthSpec struct {
	// OAuth2 - configure oauth2 authentication for the dashhboard listener
	// if configured, will be preferred over Basic authentication configuration
	// effectively disabling basic auth
	OAuth2 *DashboardOAuth2Spec `json:"oauth2,omitempty"`
	// Basic - HTTP basic auth configuration, this should only be used in exceptions
	// e.g. during evaluations or for local development
	// only used if no other authentication is configured
	Basic *DashboardBasicAuthSpec `json:"basic,omitempty"`
}

type DashboardEndpointSpec struct {
	// Auth - configure authentication for the dashboard endpoint
	Auth *DashboardAuthSpec `json:"auth,omitempty"`
	// TLS - enable and configure TLS for the Dashboard endpoint
	TLS *EndpointTlsSpec `json:"tls,omitempty"`
}

func (s *DashboardEndpointSpec) TLSSpec() *EndpointTlsSpec {
	if s == nil {
		return nil
	}

	return s.TLS
}

func (s *DashboardEndpointSpec) AuthType() DashboardAuthType {
	if s == nil || s.Auth == nil {
		return DashboardAuthTypeNone
	}

	if s.Auth.OAuth2 != nil {
		return DashboardAuthTypeOAuth2
	}

	if s.Auth.Basic != nil {
		return DashboardAuthTypeBasic
	}

	return DashboardAuthTypeNone
}

func (s *DashboardEndpointSpec) OAuth2() *DashboardOAuth2Spec {
	if s == nil || s.Auth == nil {
		return nil
	}

	return s.Auth.OAuth2
}

// APIGatewaySpec defines the desired state of APIGateway.
type APIGatewaySpec struct {
	// Envoy - configure the envoy instance and most importantly the control-plane
	Envoy *EnvoySpec `json:"envoy"`
	// ApiEndpoint - Configure the endpoint for all API routes
	// this includes the JWT configuration
	ApiEndpoint *ApiEndpointSpec `json:"apiEndpoint,omitempty"`
	// DashboardEndpoint - Configure the endpoint for the Supabase dashboard (studio)
	// this includes optional authentication (basic or Oauth2) for the dashboard
	DashboardEndpoint *DashboardEndpointSpec `json:"dashboardEndpoint,omitempty"`
	// ServiceSelector - selector to match all Supabase services (or in fact EndpointSlices) that should be considered for this APIGateway
	// +kubebuilder:default={"matchExpressions":{{"key": "app.kubernetes.io/part-of", "operator":"In", "values":{"supabase"}},{"key":"supabase.k8s.icb4dc0.de/api-gateway-target","operator":"Exists"}}}
	ServiceSelector *metav1.LabelSelector `json:"serviceSelector"`
	// ComponentTypeLabel - Label to identify which Supabase component a Service represents (e.g. auth, postgrest, ...)
	// +kubebuilder:default="app.kubernetes.io/name"
	ComponentTypeLabel string `json:"componentTypeLabel,omitempty"`
}

type EnvoyStatus struct {
	ResourceHash []byte `json:"resourceHash,omitempty"`
}

// APIGatewayStatus defines the observed state of APIGateway.
type APIGatewayStatus struct {
	Envoy          EnvoyStatus         `json:"envoy,omitempty"`
	ServiceTargets map[string][]string `json:"serviceTargets,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// APIGateway is the Schema for the apigateways API.
// +kubebuilder:printcolumn:name="EnvoyConfigVersion",type=string,JSONPath=`.status.envoy.resourceHash`
type APIGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIGatewaySpec   `json:"spec,omitempty"`
	Status APIGatewayStatus `json:"status,omitempty"`
}

func (g APIGateway) JwksSecretMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      g.Spec.ApiEndpoint.JWKSSelector.Name,
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
