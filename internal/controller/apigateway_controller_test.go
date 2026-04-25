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

package controller

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	supabasev1alpha1 "github.com/prskr/supabase-operator/api/v1alpha1"
	"github.com/prskr/supabase-operator/internal/certs"
)

var _ = Describe("APIGateway Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		apigateway := &supabasev1alpha1.APIGateway{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind APIGateway")
			err := k8sClient.Get(ctx, typeNamespacedName, apigateway)
			if err != nil && errors.IsNotFound(err) {
				resource := &supabasev1alpha1.APIGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: supabasev1alpha1.APIGatewaySpec{
						Envoy: &supabasev1alpha1.EnvoySpec{
							DisableIPv6: true,
							WorkloadSpec: &supabasev1alpha1.WorkloadSpec{
								Replicas: func() *int32 { i := int32(1); return &i }(),
							},
							ControlPlane: &supabasev1alpha1.ControlPlaneSpec{
								Host: "supabase-control-plane.supabase-system.svc",
								Port: 18000,
							},
						},
						ApiEndpoint: &supabasev1alpha1.ApiEndpointSpec{
							JWKSSelector: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "core-sample-jwt",
								},
								Key: "jwks.json",
							},
						},
						DashboardEndpoint: &supabasev1alpha1.DashboardEndpointSpec{
							TLS: &supabasev1alpha1.EndpointTlsSpec{
								Cert: &supabasev1alpha1.TlsCertRef{
									SecretName: "dashboard-tls-cert",
								},
							},
							Auth: &supabasev1alpha1.DashboardAuthSpec{
								OAuth2: &supabasev1alpha1.DashboardOAuth2Spec{
									OpenIDIssuer: "http://studio-idp:5556",
									ClientID:     "3a33e68a-c30d-4110-b4e9-bf5d45ab2126",
									Scopes:       []string{"openid", "profile", "email"},
									ClientSecretRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "studio-sample-oauth2",
										},
										Key: "clientSecret",
									},
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

			By("creating the JWKS secret")
			jwksSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "core-sample-jwt",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"jwks.json": []byte(`{"keys":[]}`),
				},
			}
			err = k8sClient.Create(ctx, jwksSecret)
			if err != nil && !errors.IsAlreadyExists(err) {
				Expect(err).NotTo(HaveOccurred())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &supabasev1alpha1.APIGateway{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance APIGateway")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Cleanup the JWKS secret")
			jwksSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "core-sample-jwt",
					Namespace: "default",
				},
			}
			_ = k8sClient.Delete(ctx, jwksSecret)
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			caCert, err := generateCACert()
			Expect(err).NotTo(HaveOccurred())

			controllerReconciler := &APIGatewayReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
				CACert: caCert,
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})

func generateCACert() (tls.Certificate, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
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

	pemCert := certs.EncodePublicKeyToPEM(derBytes)

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}
	pemKey := certs.EncodePrivateKeyToPEM(privBytes)

	return tls.X509KeyPair(pemCert, pemKey)
}
