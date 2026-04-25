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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"
)

func ServerCert(
	commonName string,
	dnsNames []string,
	ca tls.Certificate,
) (result CertResult, err error) {
	serial, err := generateSerialNumber()
	if err != nil {
		return result, err
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return result, fmt.Errorf("failed to generate private key: %w", err)
	}

	certTemplate := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: commonName,
		},
		DNSNames:    dnsNames,
		NotBefore:   time.Now().Add(-1 * time.Hour),
		NotAfter:    time.Now().AddDate(0, 1, 0),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment | x509.KeyUsageKeyEncipherment | x509.KeyUsageContentCommitment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	var caCrt *x509.Certificate
	if caCrt, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		return result, fmt.Errorf("failed parse CA cert: %w", err)
	}

	var derBytes []byte
	if derBytes, err = x509.CreateCertificate(rand.Reader, &certTemplate, caCrt, &privateKey.PublicKey, ca.PrivateKey); err != nil {
		return result, fmt.Errorf("failed to create signed certificate: %w", err)
	}

	var privateKeyBytes []byte
	if privateKeyBytes, err = x509.MarshalPKCS8PrivateKey(privateKey); err != nil {
		return result, fmt.Errorf("failed to marshal private key: %w", err)
	}

	return newCertResult(derBytes, privateKeyBytes)
}

func ClientCert(
	commonName string,
	ca tls.Certificate,
) (result CertResult, err error) {
	serial, err := generateSerialNumber()
	if err != nil {
		return result, err
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return result, fmt.Errorf("failed to generate private key: %w", err)
	}

	certTemplate := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:   time.Now().Add(-1 * time.Hour),
		NotAfter:    time.Now().AddDate(0, 0, 7),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment | x509.KeyUsageKeyEncipherment | x509.KeyUsageContentCommitment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	var caCrt *x509.Certificate
	if caCrt, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		return result, fmt.Errorf("failed parse CA cert: %w", err)
	}

	var derBytes []byte
	if derBytes, err = x509.CreateCertificate(rand.Reader, &certTemplate, caCrt, &privateKey.PublicKey, ca.PrivateKey); err != nil {
		return result, fmt.Errorf("failed to create signed certificate: %w", err)
	}

	var privateKeyBytes []byte
	if privateKeyBytes, err = x509.MarshalPKCS8PrivateKey(privateKey); err != nil {
		return result, fmt.Errorf("failed to marshal private key: %w", err)
	}

	return newCertResult(derBytes, privateKeyBytes)
}

func SelfSignedCACert() (tls.Certificate, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	serialNumber, err := generateSerialNumber()
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	pemCert := EncodePublicKeyToPEM(derBytes)

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}
	pemKey := EncodePrivateKeyToPEM(privBytes)

	return tls.X509KeyPair(pemCert, pemKey)
}

func generateSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	return rand.Int(rand.Reader, serialNumberLimit)
}
