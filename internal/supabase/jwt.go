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
	"crypto/rand"
	"encoding/hex"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type jwtDefaults struct {
	SecretKey    string
	JwksKey      string
	AnonKey      string
	ServiceKey   string
	SecretLength int
	Expiry       int
}

type jwtConfig struct {
	Defaults jwtDefaults
}

func newJwtConfig() jwtConfig {
	return jwtConfig{
		Defaults: jwtDefaults{
			SecretKey:    "secret",
			JwksKey:      "jwks.json",
			AnonKey:      "anon_key",
			ServiceKey:   "service_key",
			SecretLength: 40,
			Expiry:       3600,
		},
	}
}

func (jwtConfig) ObjectName(obj metav1.Object) string {
	return fmt.Sprintf("%s-jwt", obj.GetName())
}

func RandomJWTSecret() ([]byte, error) {
	jwtSecretBytes := make([]byte, ServiceConfig.JWT.Defaults.SecretLength)

	if _, err := rand.Read(jwtSecretBytes); err != nil {
		return nil, err
	}

	jwtSecretHex := make([]byte, hex.EncodedLen(len(jwtSecretBytes)))
	hex.Encode(jwtSecretHex, jwtSecretBytes)

	return jwtSecretHex, nil
}
