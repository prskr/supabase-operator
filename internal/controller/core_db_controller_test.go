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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	supabasev1alpha1 "github.com/prskr/supabase-operator/api/v1alpha1"
)

var _ = Describe("Core DB Controller", func() {
	Context("When reconciling a Core resource", func() {
		const resourceName = "test-resource"

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		core := new(supabasev1alpha1.Core)

		testClient := k8sClient

		BeforeEach(func(ctx SpecContext) {
			hash := fnv.New32()
			_, _ = hash.Write([]byte(ctx.SpecReport().FileName()))

			namespace := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("app-%x", hash.Sum(nil)),
				},
			}

			Expect(k8sClient.Create(ctx, namespace)).To(Succeed())

			testClient = client.NewNamespacedClient(k8sClient, namespace.Name)

			By("Preparing the environment")
			secretName := types.NamespacedName{Name: "database-credentials", Namespace: namespace.Name}
			databaseCredentialsSecret := new(v1.Secret)
			err := testClient.Get(ctx, secretName, databaseCredentialsSecret)
			if err != nil && errors.IsNotFound(err) {
				databaseCredentialsSecret = &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "database-credentials",
						Namespace: namespace.Name,
					},
					Data: map[string][]byte{
						"url": []byte("postgresql://postgres:postgres@database_host:5432/app"),
					},
				}
				Expect(testClient.Create(ctx, databaseCredentialsSecret)).To(Succeed())
			}
			By("creating the custom resource for the Kind Core")
			typeNamespacedName.Namespace = namespace.Name
			err = testClient.Get(ctx, typeNamespacedName, core)
			if err != nil && errors.IsNotFound(err) {
				resource := &supabasev1alpha1.Core{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: namespace.Name,
					},
					Spec: supabasev1alpha1.CoreSpec{
						APIExternalURL: "http://localhost:8000/",
						Database: supabasev1alpha1.Database{
							Roles: supabasev1alpha1.DatabaseRoles{
								SelfManaged: true,
								Secrets: supabasev1alpha1.DatabaseRolesSecrets{
									Admin:          "admin-role-secret",
									Authenticator:  "authenticator-role-secret",
									AuthAdmin:      "auth-admin-role-secret",
									FunctionsAdmin: "functions-admin-role-secret",
									StorageAdmin:   "storage-admin-role-secret",
								},
							},
							DSNSecretRef: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "database-credentials",
								},
								Key: "url",
							},
						},
						Postgrest: supabasev1alpha1.PostgrestSpec{
							Schemas:         []string{"public", "storage", "graphql_public"},
							ExtraSearchPath: []string{"public"},
							AnonRole:        "anon",
							MaxRows:         1000,
						},
					},
				}
				Expect(testClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &supabasev1alpha1.Core{}
			err := testClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Core")
			Expect(testClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &CoreDbReconciler{
				Client: testClient,
				Scheme: testClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})
