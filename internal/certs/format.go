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

package certs

import (
	"crypto/tls"
	"encoding/pem"
	"fmt"
)

const (
	certificateBlockType = "CERTIFICATE"
	privateKeyBlockType  = "PRIVATE KEY"
)

func EncodePublicKeyToPEM(derBytes []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: certificateBlockType, Bytes: derBytes})
}

func EncodePrivateKeyToPEM(privateKeyBytes []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: privateKeyBlockType, Bytes: privateKeyBytes})
}

func newCertResult(derBytes, privateKeyBytes []byte) (CertResult, error) {
	pemEncodedPublicKey := EncodePublicKeyToPEM(derBytes)
	pemEncodedPrivateKey := EncodePrivateKeyToPEM(privateKeyBytes)

	cert, err := tls.X509KeyPair(pemEncodedPublicKey, pemEncodedPrivateKey)
	if err != nil {
		return CertResult{}, fmt.Errorf("failed to create TLS cert based on x509 key pair: %w", err)
	}

	result := CertResult{
		ServerCert: cert,
		PublicKey:  pemEncodedPublicKey,
		PrivateKey: pemEncodedPrivateKey,
	}

	return result, nil
}

type CertResult struct {
	ServerCert tls.Certificate
	PublicKey  []byte
	PrivateKey []byte
}
