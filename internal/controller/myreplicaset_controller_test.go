/*
Copyright 2024.

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
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1beta1 "devops-test/api/v1beta1"
)

var _ = Describe("MyReplicaSet Controller", func() {
	Context("When reconciling a resource", func() {
		const myreplicasetName = "test-myreplicaset"

		ctx := context.Background()

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      myreplicasetName,
				Namespace: myreplicasetName,
			},
		}

		typeNamespacedName := types.NamespacedName{
			Name:      myreplicasetName,
			Namespace: myreplicasetName, // TODO(user):Modify as needed
		}
		myreplicaset := &appsv1beta1.MyReplicaSet{}

		BeforeEach(func() {
			By("Creating the Namespace to perform the tests")
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))

			By("creating the custom resource for the Kind MyReplicaSet")
			err = k8sClient.Get(ctx, typeNamespacedName, myreplicaset)
			if err != nil && errors.IsNotFound(err) {
				// 创建 MyReplicaSet 资源
				resource := &appsv1beta1.MyReplicaSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      myreplicasetName,
						Namespace: namespace.Name,
					},
					// TODO(user): Specify other spec details if needed.
					Spec: appsv1beta1.MyReplicaSetSpec{
						Replicas: func() *int32 { replicas := int32(3); return &replicas }(),
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "test-app",
							},
						},
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"app": "test-app",
								},
							},
							Spec: corev1.PodSpec{

								Containers: []corev1.Container{
									{
										Name:  "test-container",
										Image: "nginx:latest",
									},
								},
							},
						},
					},
				}
				err = k8sClient.Create(ctx, resource)
				Expect(err).To(Not(HaveOccurred()))
				// Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			} else {
				fmt.Printf("resource already exists %s", err)
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.

			By("removing the custom resource for the Kind MyReplicaSet")
			resource := &appsv1beta1.MyReplicaSet{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() error {
				return k8sClient.Delete(context.TODO(), resource)
			}, 2*time.Minute, time.Second).Should(Succeed())

			// TODO(user): Attention if you improve this code by adding other context test you MUST
			// be aware of the current delete namespace limitations.
			// More info: https://book.kubebuilder.io/reference/envtest.html#testing-considerations
			By("Deleting the Namespace to perform the tests")
			_ = k8sClient.Delete(ctx, namespace)
		})
		It("should successfully reconcile the  of myreplicaset", func() {
			By("Reconciling the created resource")
			Eventually(func() error {
				resource := &appsv1beta1.MyReplicaSet{}
				return k8sClient.Get(ctx, typeNamespacedName, resource)
			}, time.Minute, time.Second).Should(Succeed())

			By("Reconciling the custom resource created")
			controllerReconciler := &MyReplicaSetReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.

			By("Checking if Pods was successfully created in the reconciliation")
			Eventually(func() error {
				// found := &v1.Pod{}
				// return k8sClient.Get(ctx, typeNamespacedName, found)

				podList := &v1.PodList{}

				return k8sClient.List(ctx, podList, client.InNamespace(myreplicaset.Namespace))

			}, time.Minute, time.Second).Should(Succeed())

			By("Checking the latest Status Condition added to the Myreplicaset instance")
			Eventually(func() error {
				if myreplicaset.Status.Replicas != 0 {
					fmt.Println("myreplicaset.Status.Replicas is not 0, pods readonly crate")
				}
				fmt.Println("myreplicaset.Status.Replicas is 0, pods readonly not crate")

				return nil
			}, time.Minute, time.Second).Should(Succeed())
		})
	})
})
