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
	"fmt"
	"hash/fnv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	supabasev1alpha1 "github.com/prskr/supabase-operator/api/v1alpha1"
)

var _ = Describe("Dashboard Studio Controller", func() {
	Context("When reconciling a Dashboard resource", func() {
		const resourceName = "test-resource"

		var (
			namespace          *v1.Namespace
			namespaceName      string
			testClient         client.Client
			typeNamespacedName types.NamespacedName
		)

		BeforeEach(func(ctx SpecContext) {
			hash := fnv.New32()
			_, _ = hash.Write([]byte(ctx.SpecReport().FileName()))

			namespaceName = fmt.Sprintf("dashboard-studio-%x", hash.Sum(nil))
			namespace = &v1.Namespace{
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

			By("Creating supporting resources")

			// Create DB credentials secret
			By("Creating DB credentials secret")
			dbCredsSecret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-db-creds",
					Namespace: namespaceName,
				},
				Data: map[string][]byte{
					"username": []byte("postgres"),
					"password": []byte("postgres-password"),
				},
			}
			Expect(testClient.Create(ctx, dbCredsSecret)).To(Succeed())

			// Create JWT secret
			By("Creating JWT secret")
			jwtSecret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-jwt",
					Namespace: namespaceName,
				},
				Data: map[string][]byte{
					"secret":      []byte("test-jwt-secret-key-12345678"),
					"jwks.json":   []byte(`{"keys":[]}`),
					"anon_key":    []byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.anon"),
					"service_key": []byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.service"),
				},
			}
			Expect(testClient.Create(ctx, jwtSecret)).To(Succeed())

			// Create gateway service that matches the default selector
			By("Creating API gateway service")
			gatewayService := &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "api-gateway",
					Namespace: namespaceName,
					Labels: map[string]string{
						"app.kubernetes.io/name":      "envoy",
						"app.kubernetes.io/component": "api-gateway",
					},
				},
				Spec: v1.ServiceSpec{
					Selector: map[string]string{
						"app": "envoy",
					},
					Ports: []v1.ServicePort{
						{
							Port:       8000,
							TargetPort: intstr.FromInt(8000),
						},
					},
				},
			}
			Expect(testClient.Create(ctx, gatewayService)).To(Succeed())

			// Create the Dashboard resource
			By("Creating the Dashboard resource")
			dashboard := &supabasev1alpha1.Dashboard{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespaceName,
				},
				Spec: supabasev1alpha1.DashboardSpec{
					DBSpec: &supabasev1alpha1.DashboardDBSpec{
						Host:   "cluster-example-rw.supabase-demo.svc",
						Port:   5432,
						DBName: "app",
						DBCredentialsRef: &supabasev1alpha1.DBCredentialsReference{
							SecretName: "test-db-creds",
						},
						Schemas:         []string{"public", "storage", "graphql_public"},
						ExtraSearchPath: []string{"public"},
					},
					Studio: &supabasev1alpha1.StudioSpec{
						APIExternalURL: "http://localhost:8000",
						JWT: &supabasev1alpha1.JwtSpec{
							SecretName: "test-jwt",
						},
					},
				},
			}
			Expect(testClient.Create(ctx, dashboard)).To(Succeed())
		})

		AfterEach(func(ctx SpecContext) {
			By("Cleaning up resources")
			dashboard := &supabasev1alpha1.Dashboard{}
			err := testClient.Get(ctx, typeNamespacedName, dashboard)
			if err == nil || !errors.IsNotFound(err) {
				Expect(testClient.Delete(ctx, dashboard)).To(Succeed())
			}

			// Delete namespace to clean up all resources
			Expect(testClient.Delete(ctx, namespace)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &DashboardStudioReconciler{
				Client: testClient,
				Scheme: testClient.Scheme(),
			}

			result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeZero())

			// Verify the Studio Deployment was created
			By("Verifying Studio Deployment was created")
			studioDeployment := &appsv1.Deployment{}
			studioDeploymentName := types.NamespacedName{
				Name:      fmt.Sprintf("%s-studio", resourceName),
				Namespace: namespaceName,
			}
			Eventually(func() bool {
				err := testClient.Get(ctx, studioDeploymentName, studioDeployment)
				return err == nil
			}).Should(BeTrue(), "Studio Deployment should be created")

			Expect(studioDeployment.Name).To(Equal(studioDeploymentName.Name))
			Expect(studioDeployment.Namespace).To(Equal(studioDeploymentName.Namespace))
			Expect(studioDeployment.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(studioDeployment.Spec.Template.Spec.Containers[0].Name).To(Equal("supabase-studio"))

			// Verify the Studio Service was created
			By("Verifying Studio Service was created")
			studioService := &v1.Service{}
			studioServiceName := types.NamespacedName{
				Name:      fmt.Sprintf("%s-studio", resourceName),
				Namespace: namespaceName,
			}
			Eventually(func() bool {
				err := testClient.Get(ctx, studioServiceName, studioService)
				return err == nil
			}).Should(BeTrue(), "Studio Service should be created")

			Expect(studioService.Name).To(Equal(studioServiceName.Name))
			Expect(studioService.Namespace).To(Equal(studioServiceName.Namespace))
			Expect(studioService.Spec.Ports).To(HaveLen(1))
			Expect(studioService.Spec.Ports[0].Port).To(Equal(int32(3000)))
		})
	})
})
