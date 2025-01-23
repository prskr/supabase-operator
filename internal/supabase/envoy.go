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
			ConfigKey: "config.yaml",
			UID:       65532,
			GID:       65532,
		},
	}
}

type envoyDefaults struct {
	ConfigKey string
	UID, GID  int64
}

type envoyServiceConfig struct {
	Defaults envoyDefaults
}

func (envoyServiceConfig) ObjectName(obj metav1.Object) string {
	return fmt.Sprintf("%s-envoy", obj.GetName())
}
