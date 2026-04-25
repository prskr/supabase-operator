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
	"crypto/rand"
	"fmt"
	"hash/fnv"
	"time"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/gkampitakis/go-snaps/match"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	supabasev1alpha1 "github.com/prskr/supabase-operator/api/v1alpha1"
	"github.com/prskr/supabase-operator/internal/certs"
	"github.com/prskr/supabase-operator/internal/supabase"
)

var _ = Describe("APIGateway Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		var (
			namespace            *corev1.Namespace
			namespaceName        string
			testClient           client.Client
			typeNamespacedName   types.NamespacedName
			apigateway           *supabasev1alpha1.APIGateway
			reconciliationErr    error
			reconciliationResult reconcile.Result
		)

		BeforeEach(func(ctx SpecContext) {
			hash := fnv.New32()
			_, _ = hash.Write([]byte(ctx.SpecReport().FullText()))

			randomSuffix := make([]byte, 4)
			_, _ = rand.Read(randomSuffix)

			namespaceName = fmt.Sprintf("apigateway-%x-%x", hash.Sum(nil), randomSuffix)
			namespace = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespaceName,
				},
			}

			Expect(k8sClient.Create(ctx, namespace)).To(Succeed())
			testClient = client.NewNamespacedClient(k8sClient, namespaceName)

			typeNamespacedName = types.NamespacedName{
				Name:      resourceName,
				Namespace: namespaceName,
			}

			By("creating the custom resource for the Kind APIGateway")
			apigateway = &supabasev1alpha1.APIGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespaceName,
				},
				Spec: supabasev1alpha1.APIGatewaySpec{
					Envoy: &supabasev1alpha1.EnvoySpec{
						DisableIPv6: true,
						WorkloadSpec: &supabasev1alpha1.WorkloadSpec{
							Replicas: new(int32(1)),
						},
						ControlPlane: &supabasev1alpha1.ControlPlaneSpec{
							Host: "supabase-control-plane.supabase-system.svc",
							Port: 18000,
						},
						Observability: &supabasev1alpha1.EnvoyObservabilitySpec{
							Metrics: &supabasev1alpha1.EnvoyMetricsSpec{
								Enabled: true,
							},
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
			Expect(testClient.Create(ctx, apigateway)).To(Succeed())

			By("creating the JWKS secret")
			jwksSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "core-sample-jwt",
					Namespace: namespaceName,
				},
				Data: map[string][]byte{
					"jwks.json": []byte(`{"keys":[]}`),
				},
			}
			Expect(testClient.Create(ctx, jwksSecret)).To(Succeed())
		})

		AfterEach(func(ctx SpecContext) {
			By("Cleaning up resources")
			resource := &supabasev1alpha1.APIGateway{}
			err := testClient.Get(ctx, typeNamespacedName, resource)
			if err == nil || !errors.IsNotFound(err) {
				Expect(testClient.Delete(ctx, resource)).To(Succeed())
			}

			Expect(k8sClient.Delete(ctx, namespace)).To(Succeed())
		})

		JustBeforeEach(func(ctx SpecContext) {
			By("Reconciling the created resource")
			caCert, err := certs.SelfSignedCACert()
			Expect(err).NotTo(HaveOccurred())

			controllerReconciler := &APIGatewayReconciler{
				Client: testClient,
				Scheme: testClient.Scheme(),
				CACert: caCert,
			}

			reconciliationResult, reconciliationErr = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
		})

		It("should successfully reconcile the resource", func() {
			Expect(reconciliationErr).NotTo(HaveOccurred())
			Expect(reconciliationResult).To(Equal(reconcile.Result{RequeueAfter: 15 * time.Minute}))
		})

		It("should successfully create the HMAC secret", func(ctx SpecContext) {
			By("Checking the created HMAC secret")
			err := testClient.Get(ctx, typeNamespacedName, apigateway)
			Expect(err).NotTo(HaveOccurred())

			hmacSecret := &corev1.Secret{}
			hmacSecretName := supabase.ServiceConfig.Envoy.HmacSecretName(apigateway)
			err = testClient.Get(ctx, types.NamespacedName{Name: hmacSecretName, Namespace: namespaceName}, hmacSecret)
			Expect(err).NotTo(HaveOccurred())
			Expect(hmacSecret.Data).To(HaveKey(supabase.ServiceConfig.Envoy.Defaults.HmacSecretKey))

			snapshotConfig.MatchJSON(GinkgoT(), hmacSecret, match.Any(
				"data.oauth2_hmac_secret",
				"metadata.resourceVersion",
				"metadata.creationTimestamp",
				"metadata.uid",
				"metadata.managedFields",
				"metadata.namespace",
				"metadata.ownerReferences.0.uid",
			))
		})

		It("should successfully create the client certificate secret", func(ctx SpecContext) {
			By("Checking the created client certificate secret")
			err := testClient.Get(ctx, typeNamespacedName, apigateway)
			Expect(err).NotTo(HaveOccurred())

			clientCertSecret := &corev1.Secret{}
			clientCertSecretName := supabase.ServiceConfig.Envoy.ControlPlaneClientCertSecretName(apigateway)
			err = testClient.Get(ctx, types.NamespacedName{Name: clientCertSecretName, Namespace: namespaceName}, clientCertSecret)
			Expect(err).NotTo(HaveOccurred())
			Expect(clientCertSecret.Data).To(HaveKey("ca.crt"))
			Expect(clientCertSecret.Data).To(HaveKey(corev1.TLSCertKey))
			Expect(clientCertSecret.Data).To(HaveKey(corev1.TLSPrivateKeyKey))

			snapshotConfig.MatchJSON(GinkgoT(), clientCertSecret, match.Any(
				"data.ca\\.crt",
				"data.tls\\.crt",
				"data.tls\\.key",
				"metadata.resourceVersion",
				"metadata.creationTimestamp",
				"metadata.uid",
				"metadata.managedFields",
				"metadata.namespace",
				"metadata.ownerReferences.0.uid",
			))
		})

		It("should successfully create the Envoy ConfigMap", func(ctx SpecContext) {
			By("Checking the created Envoy ConfigMap")
			err := testClient.Get(ctx, typeNamespacedName, apigateway)
			Expect(err).NotTo(HaveOccurred())

			configMap := &corev1.ConfigMap{}
			configMapName := supabase.ServiceConfig.Envoy.ObjectName(apigateway)
			err = testClient.Get(ctx, types.NamespacedName{Name: configMapName, Namespace: namespaceName}, configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap.Data).To(HaveKey(supabase.ServiceConfig.Envoy.Defaults.ConfigKey))

			snapshotConfig.MatchJSON(GinkgoT(), configMap, match.Any(
				"metadata.resourceVersion",
				"metadata.creationTimestamp",
				"metadata.uid",
				"metadata.managedFields",
				"metadata.namespace",
				"metadata.ownerReferences.0.uid",
				"data.config\\.yaml",
			))
		})

		It("should successfully create the Envoy Deployment", func(ctx SpecContext) {
			By("Checking the created Envoy Deployment")
			err := testClient.Get(ctx, typeNamespacedName, apigateway)
			Expect(err).NotTo(HaveOccurred())

			deployment := &appsv1.Deployment{}
			deploymentName := supabase.ServiceConfig.Envoy.ObjectName(apigateway)
			err = testClient.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: namespaceName}, deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(deployment.Spec.Template.Spec.Containers[0].Name).To(Equal("envoy-proxy"))

			snapshotConfig.MatchJSON(GinkgoT(), deployment, match.Any(
				"metadata.resourceVersion",
				"metadata.creationTimestamp",
				"metadata.uid",
				"metadata.managedFields",
				"metadata.namespace",
				"metadata.ownerReferences.0.uid",
				"spec.template.metadata.annotations.supabase\\.k8s\\.icb4dc0\\.de/config-hash",
				"spec.template.metadata.annotations.supabase\\.k8s\\.icb4dc0\\.de/jwks-hash",
			))
		})

		It("should successfully create the Envoy Service", func(ctx SpecContext) {
			By("Checking the created Envoy Service")
			err := testClient.Get(ctx, typeNamespacedName, apigateway)
			Expect(err).NotTo(HaveOccurred())

			service := &corev1.Service{}
			serviceName := supabase.ServiceConfig.Envoy.ObjectName(apigateway)
			err = testClient.Get(ctx, types.NamespacedName{Name: serviceName, Namespace: namespaceName}, service)
			Expect(err).NotTo(HaveOccurred())

			snapshotConfig.MatchJSON(GinkgoT(), service, match.Any(
				"metadata.resourceVersion",
				"metadata.creationTimestamp",
				"metadata.uid",
				"metadata.managedFields",
				"metadata.namespace",
				"metadata.ownerReferences.0.uid",
				"spec.clusterIP",
				"spec.clusterIPs",
			))
		})

		It("should successfully create the Envoy PodMonitor", func(ctx SpecContext) {
			By("Checking the created Envoy PodMonitor")
			err := testClient.Get(ctx, typeNamespacedName, apigateway)
			Expect(err).NotTo(HaveOccurred())

			podMonitor := &monitoringv1.PodMonitor{}
			podMonitorName := supabase.ServiceConfig.Envoy.ObjectName(apigateway)
			err = testClient.Get(ctx, types.NamespacedName{Name: podMonitorName, Namespace: namespaceName}, podMonitor)
			Expect(err).NotTo(HaveOccurred())

			snapshotConfig.MatchJSON(GinkgoT(), podMonitor, match.Any(
				"metadata.resourceVersion",
				"metadata.creationTimestamp",
				"metadata.uid",
				"metadata.managedFields",
				"metadata.namespace",
				"metadata.ownerReferences.0.uid",
			))
		})
	})
})
