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

package supabase

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newEnvoyServiceConfig() envoyServiceConfig {
	return envoyServiceConfig{
		Defaults: envoyDefaults{
			ConfigKey:             "config.yaml",
			OAuth2ClientSecretKey: "oauth2_client_secret",
			HmacSecretKey:         "oauth2_hmac_secret",
			UID:                   65532,
			GID:                   65532,
			StudioPortName:        "studio",
			ApiPortName:           "api",
			StudioPort:            3000,
			ApiPort:               8000,
			AdminPort:             19000,
		},
	}
}

type envoyDefaults struct {
	ConfigKey                      string
	HmacSecretKey                  string
	OAuth2ClientSecretKey          string
	UID, GID                       int64
	StudioPortName, ApiPortName    string
	StudioPort, ApiPort, AdminPort int32
}

type envoyServiceConfig struct {
	Defaults envoyDefaults
}

func (envoyServiceConfig) ObjectName(obj metav1.Object) string {
	return fmt.Sprintf("%s-envoy", obj.GetName())
}

func (envoyServiceConfig) ControlPlaneClientCertSecretName(obj metav1.Object) string {
	return fmt.Sprintf("%s-cp-client-cert", obj.GetName())
}

func (envoyServiceConfig) HmacSecretName(obj metav1.Object) string {
	return fmt.Sprintf("%s-hmac-secret", obj.GetName())
}
