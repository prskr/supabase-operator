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

type serviceConfig[TEnvKeys, TDefaults any] struct {
	Name               string
	LivenessProbePath  string
	ReadinessProbePath string
	EnvKeys            TEnvKeys
	Defaults           TDefaults
}

func (cfg serviceConfig[TEnvKeys, TDefaults]) ReadinessPath() string {
	return cfg.ReadinessProbePath
}

func (cfg serviceConfig[TEnvKeys, TDefaults]) LivenessPath() string {
	if cfg.LivenessProbePath == "" {
		return cfg.ReadinessProbePath
	}
	return cfg.LivenessProbePath
}

func (cfg serviceConfig[TEnvKeys, TDefaults]) ObjectName(obj metav1.Object) string {
	return fmt.Sprintf("%s-%s", obj.GetName(), cfg.Name)
}

func (cfg serviceConfig[TEnvKeys, TDefaults]) ObjectMeta(obj metav1.Object) metav1.ObjectMeta {
	return metav1.ObjectMeta{Name: cfg.ObjectName(obj), Namespace: obj.GetNamespace()}
}

var ServiceConfig = struct {
	Postgrest serviceConfig[postgrestEnvKeys, postgrestConfigDefaults]
	Auth      serviceConfig[authEnvKeys, authConfigDefaults]
	PGMeta    serviceConfig[pgMetaEnvKeys, pgMetaDefaults]
	Studio    serviceConfig[studioEnvKeys, studioDefaults]
	Storage   serviceConfig[storageEnvApiKeys, storageApiDefaults]
	ImgProxy  serviceConfig[imgProxyEnvKeys, imgProxyDefaults]
	Envoy     envoyServiceConfig
	JWT       jwtConfig
}{
	Postgrest: postgrestServiceConfig(),
	Auth:      authServiceConfig(),
	PGMeta:    pgMetaServiceConfig(),
	Studio:    studioServiceConfig(),
	Storage:   storageServiceConfig(),
	ImgProxy:  imgProxyServiceConfig(),
	Envoy:     newEnvoyServiceConfig(),
	JWT:       newJwtConfig(),
}
