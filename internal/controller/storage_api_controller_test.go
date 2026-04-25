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

	supabasev1alpha1 "github.com/prskr/supabase-operator/api/v1alpha1"
	"github.com/prskr/supabase-operator/internal/supabase"

	"github.com/gkampitakis/go-snaps/match"
	"github.com/gkampitakis/go-snaps/snaps"
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
)

var _ = Describe("Storage API Controller", func() {
	Context("When reconciling a Storage resource", func() {
		const resourceName = "test-resource"

		var (
			namespace            *v1.Namespace
			namespaceName        string
			testClient           client.Client
			typeNamespacedName   types.NamespacedName
			storage              *supabasev1alpha1.Storage
			reconciliationErr    error
			reconciliationResult reconcile.Result
		)

		BeforeEach(func(ctx SpecContext) {
			hash := fnv.New32()
			_, _ = hash.Write([]byte(ctx.SpecReport().FullText()))

			randomSuffix := make([]byte, 4)
			_, _ = rand.Read(randomSuffix)

			namespaceName = fmt.Sprintf("storage-api-%x-%x", hash.Sum(nil), randomSuffix)
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
					"username": []byte("supabase_storage_admin"),
					"password": []byte("test-db-password"),
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
					"anon_key":    []byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS10ZXN0Iiwicm9sZSI6ImFub24iLCJleHAiOjE3MDAwMDAwMDB9.anon"),
					"service_key": []byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS10ZXN0Iiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImV4cCI6MTcwMDAwMDAwMH0.service"),
				},
			}
			Expect(testClient.Create(ctx, jwtSecret)).To(Succeed())

			// Create S3 protocol credentials secret
			By("Creating S3 protocol credentials secret")
			s3ProtocolCredentialsSecret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-storage-protocol-s3-credentials", resourceName),
					Namespace: namespaceName,
				},
				StringData: map[string]string{
					"accessKeyId":     "test-access-key-id",
					"secretAccessKey": "test-secret-access-key",
				},
			}
			Expect(testClient.Create(ctx, s3ProtocolCredentialsSecret)).To(Succeed())

			// Create S3 backend credentials secret
			By("Creating S3 backend credentials secret")
			s3BackendCredentialsSecret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-s3-backend-credentials",
					Namespace: namespaceName,
				},
				StringData: map[string]string{
					"accessKeyId":     "test-access-key-id",
					"secretAccessKey": "test-secret-access-key",
				},
			}
			Expect(testClient.Create(ctx, s3BackendCredentialsSecret)).To(Succeed())

			// Create PostgREST service that matches the default selector
			By("Creating PostgREST service")
			postgrestService := &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "postgrest",
					Namespace: namespaceName,
					Labels: map[string]string{
						"app.kubernetes.io/name":      "postgrest",
						"app.kubernetes.io/component": "core",
					},
				},
				Spec: v1.ServiceSpec{
					Selector: map[string]string{
						"app": "postgrest",
					},
					Ports: []v1.ServicePort{
						{
							Name:       "api",
							Port:       3000,
							TargetPort: intstr.FromInt(3000),
						},
					},
				},
			}
			Expect(testClient.Create(ctx, postgrestService)).To(Succeed())

			// Create the Storage resource
			By("Creating the Storage resource")
			storage = &supabasev1alpha1.Storage{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespaceName,
				},
				Spec: supabasev1alpha1.StorageSpec{
					ImageProxy: &supabasev1alpha1.ImageProxySpec{
						Enable: true,
					},
					API: supabasev1alpha1.StorageAPISpec{
						TenantID: new("test-tenant-id"),
						Region:   new("us-east-1"),
						S3Backend: &supabasev1alpha1.S3BackendSpec{
							Endpoint: "http://seaweedfs.seaweedfs.svc:8333",
							Region:   "us-east-1",
							Bucket:   "test-bucket",
							CredentialsSecretRef: &supabasev1alpha1.S3CredentialsRef{
								SecretName: "test-s3-backend-credentials",
							},
						},
						S3Protocol: &supabasev1alpha1.S3ProtocolSpec{
							AllowForwardedHeader: true,
							CredentialsSecretRef: &supabasev1alpha1.S3CredentialsRef{
								SecretName: fmt.Sprintf("%s-storage-protocol-s3-credentials", resourceName),
							},
						},
						JwtAuth: supabasev1alpha1.JwtSpec{
							SecretName: "test-jwt",
							SecretKey:  "secret",
							JwksKey:    "jwks.json",
							AnonKey:    "anon_key",
							ServiceKey: "service_key",
						},
						DBSpec: supabasev1alpha1.StorageAPIDBSpec{
							Host:   "cluster-test-rw.supabase-demo.svc",
							Port:   5432,
							DBName: "app",
							DBCredentialsRef: &supabasev1alpha1.DBCredentialsReference{
								SecretName: "test-db-creds",
							},
						},
					},
				},
			}
			Expect(testClient.Create(ctx, storage)).To(Succeed())
		})

		AfterEach(func(ctx SpecContext) {
			By("Cleaning up resources")
			storage := &supabasev1alpha1.Storage{}
			err := testClient.Get(ctx, typeNamespacedName, storage)
			if err == nil || !errors.IsNotFound(err) {
				Expect(testClient.Delete(ctx, storage)).To(Succeed())
			}

			// Delete namespace to clean up all resources
			Expect(testClient.Delete(ctx, namespace)).To(Succeed())
		})

		JustBeforeEach(func(ctx SpecContext) {
			By("Reconciling the created resource")
			controllerReconciler := &StorageAPIReconciler{
				Client: testClient,
				Scheme: testClient.Scheme(),
			}

			reconciliationResult, reconciliationErr = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
		})

		It("should successfully reconcile the resource", func() {
			Expect(reconciliationErr).NotTo(HaveOccurred())
			Expect(reconciliationResult).To(BeZero())
		})

		It("should create the Storage API Deployment", func(ctx SpecContext) {
			By("Verifying Storage API Deployment was created")
			storageDeployment := &appsv1.Deployment{}
			deploymentName := types.NamespacedName{
				Name:      supabase.ServiceConfig.Storage.ObjectName(storage),
				Namespace: namespaceName,
			}
			Eventually(func() error {
				return testClient.Get(ctx, deploymentName, storageDeployment)
			}).Should(Succeed(), "Storage API Deployment should be created")

			Expect(storageDeployment.Name).To(Equal(deploymentName.Name))
			Expect(storageDeployment.Namespace).To(Equal(deploymentName.Namespace))
			Expect(storageDeployment.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(storageDeployment.Spec.Template.Spec.Containers[0].Name).To(Equal("supabase-storage"))

			snaps.MatchJSON(GinkgoT(), storageDeployment, match.Any(
				"metadata.resourceVersion",
				"metadata.creationTimestamp",
				"metadata.uid",
				"metadata.managedFields",
				"metadata.ownerReferences.0.uid",
				"spec.template.metadata.annotations.supabase\\.k8s\\.icb4dc0\\.de/jwt-hash",
				"spec.template.metadata.annotations.supabase\\.k8s\\.icb4dc0\\.de/s3-credentials-hash",
			))
		})

		It("should create the Storage API Service", func(ctx SpecContext) {
			By("Verifying Storage API Service was created")
			storageService := &v1.Service{}
			serviceName := types.NamespacedName{
				Name:      supabase.ServiceConfig.Storage.ObjectName(storage),
				Namespace: namespaceName,
			}
			Eventually(func() error {
				return testClient.Get(ctx, serviceName, storageService)
			}).Should(Succeed(), "Storage API Service should be created")

			Expect(storageService.Name).To(Equal(serviceName.Name))
			Expect(storageService.Namespace).To(Equal(serviceName.Namespace))
			Expect(storageService.Spec.Ports).To(HaveLen(1))
			Expect(storageService.Spec.Ports[0].Port).To(Equal(int32(5000)))

			snaps.MatchJSON(GinkgoT(), storageService, match.Any(
				"metadata.resourceVersion",
				"metadata.creationTimestamp",
				"metadata.uid",
				"metadata.managedFields",
				"metadata.ownerReferences.0.uid",
			))
		})
	})
})
