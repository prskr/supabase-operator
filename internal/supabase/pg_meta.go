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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type pgMetaEnvKeys struct {
	APIPort    intEnv[int32]
	DBHost     stringEnv
	DBPort     intEnv[int]
	DBName     stringEnv
	DBUser     secretEnv
	DBPassword secretEnv
	CryptoKey  secretEnv
}

type pgMetaDefaults struct {
	APIPort         int32
	DBPort          string
	NodeUID         int64
	NodeGID         int64
	CryptoKeyKey    string
	CryptoKeyLength int
}

type pgMetaConfig struct {
	serviceConfig[pgMetaEnvKeys, pgMetaDefaults]
}

func (c pgMetaConfig) CryptoKeySecretName(obj metav1.Object) string {
	return c.ObjectName(obj) + "-crypto"
}

func (c pgMetaConfig) CryptoKeySelector(obj metav1.Object) *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: c.CryptoKeySecretName(obj),
		},
		Key: c.Defaults.CryptoKeyKey,
	}
}

func pgMetaServiceConfig() pgMetaConfig {
	return pgMetaConfig{
		serviceConfig: serviceConfig[pgMetaEnvKeys, pgMetaDefaults]{
			Name:              "pg-meta",
			LivenessProbePath: "/health",
			EnvKeys: pgMetaEnvKeys{
				APIPort:    "PG_META_PORT",
				DBHost:     "PG_META_DB_HOST",
				DBPort:     "PG_META_DB_PORT",
				DBName:     "PG_META_DB_NAME",
				DBUser:     "PG_META_DB_USER",
				DBPassword: "PG_META_DB_PASSWORD",
				CryptoKey:  "PG_META_CRYPTO_KEY",
			},
			Defaults: pgMetaDefaults{
				APIPort:         8080,
				DBPort:          "5432",
				NodeUID:         1000,
				NodeGID:         1000,
				CryptoKeyKey:    "key",
				CryptoKeyLength: 32,
			},
		},
	}
}
