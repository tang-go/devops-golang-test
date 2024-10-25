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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	myappv1 "hzy.com/mystate/api/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

var _ = Describe("MyState Controller", func() {
	const (
		resourceName = "test-resource"
		timeout      = time.Second * 5
		interval     = time.Millisecond * 250
	)

	var ctx context.Context

	typeNamespacedName := types.NamespacedName{
		Name:      resourceName,
		Namespace: "default", // TODO(user):Modify as needed
	}
	var mystate *myappv1.MyState

	BeforeEach(func() {
		mystate = &myappv1.MyState{
			ObjectMeta: metav1.ObjectMeta{
				Name:      resourceName,
				Namespace: "default",
			},
			Spec: myappv1.MyStateSpec{
				Replicas: 1,
				Selector: &metav1.LabelSelector{
					MatchLabels:      nil,
					MatchExpressions: nil,
				},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "test-container",
								Image: "test-image",
							},
						},
					},
				},
				VolumeClaimTemplates: nil,
				ServiceName:          "",
				Ordinals:             myappv1.MyOrdinals{},
			},
		}
		ctx = context.Background()
	})

	AfterEach(func() {
		// TODO(user): Cleanup logic after each test, like removing the resource instance.
		//resource := &myappv1.MyState{}
		//err := k8sClient.Get(ctx, typeNamespacedName, resource)
		//Expect(err).NotTo(HaveOccurred())
		//
		//By("Cleanup the specific resource instance MyState")
		//Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		//
		//var childPods v1.PodList
		//
		//err = k8sClient.List(ctx, &childPods)
		//Expect(err).To(BeNil())
		//
		//for _, pod := range childPods.Items {
		//	err = k8sClient.Delete(ctx, &pod)
		//	Expect(err).To(BeNil())
		//}
	})
	Context("When reconciling a resource", func() {

		It("Should add finalizer", func() {
			Expect(k8sClient.Create(ctx, mystate)).To(Succeed())

			tryGet := &myappv1.MyState{}
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, typeNamespacedName, tryGet)).To(Succeed())
			}, timeout, interval).Should(Succeed())

			Eventually(func(g Gomega) {
				k8sClient.Get(ctx, typeNamespacedName, tryGet)
				g.Expect(len(tryGet.GetFinalizers())).To(Equal(1))
			}, timeout, interval).Should(Succeed())

			By("create a pod")

			//pod, err := controller.generatePod(*mystate, 0)
			//
			//Expect(err).To(BeNil())
			//Expect(k8sClient.Create(ctx, &pod)).To(Succeed())

			Eventually(func(g Gomega) {
				var childPods v1.PodList
				err := k8sClient.List(ctx, &childPods)
				g.Expect(err).To(BeNil())
				g.Expect(len(childPods.Items)).To(Equal(1))
			}, timeout, interval).Should(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, typeNamespacedName, tryGet)).To(Succeed())
				g.Expect(tryGet.Status.Replicas).To(Equal(1))
			}, timeout, interval).Should(Succeed())

			By("reduce replicas")

			tryGet.Spec.Replicas = 0
			Expect(k8sClient.Update(ctx, tryGet)).To(Succeed())
			//var childPods v1.PodList
			//err := k8sClient.List(ctx, &childPods)
			//Expect(err).To(BeNil())
			//Expect(len(childPods.Items)).To(Equal(1))
			//for _, pod := range childPods.Items {
			//	Expect(k8sClient.Delete(ctx, &pod)).To(Succeed())
			//}

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, typeNamespacedName, tryGet)).To(Succeed())
				g.Expect(tryGet.Status.Replicas).To(Equal(0))
			}, timeout, interval).Should(Succeed())

		})
		//
		//Context("already have finalizer", func() {
		//	It("Should add first pod", func() {
		//		controllerReconciler := &MyStateReconciler{
		//			Client: k8sClient,
		//			Scheme: k8sClient.Scheme(),
		//		}
		//
		//		newState := mystate.DeepCopy()
		//		controllerReconciler.addFinalizer(newState)
		//		err = k8sClient.Update(ctx, newState)
		//		Expect(err).To(BeNil())
		//
		//		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
		//			NamespacedName: typeNamespacedName,
		//		})
		//		Expect(err).To(BeNil())
		//
		//		var childPods v1.PodList
		//
		//		err = k8sClient.List(ctx, &childPods, client.InNamespace(newState.Namespace), client.MatchingFields{podOwnerKey: newState.Name})
		//		Expect(err).To(BeNil())
		//		Expect(len(childPods.Items)).To(Equal(1))
		//	})
		//})
		//
		//Context("already have finalizer", func() {
		//	It("Should add first pod", func() {
		//		controllerReconciler := &MyStateReconciler{
		//			Client: k8sClient,
		//			Scheme: k8sClient.Scheme(),
		//		}
		//
		//		newState := mystate.DeepCopy()
		//		controllerReconciler.addFinalizer(newState)
		//		err = k8sClient.Update(ctx, newState)
		//		Expect(err).To(BeNil())
		//
		//		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
		//			NamespacedName: typeNamespacedName,
		//		})
		//		Expect(err).To(BeNil())
		//
		//		var childPods v1.PodList
		//
		//		err = k8sClient.List(ctx, &childPods, client.InNamespace(newState.Namespace), client.MatchingFields{podOwnerKey: newState.Name})
		//		Expect(err).To(BeNil())
		//		Expect(len(childPods.Items)).To(Equal(1))
		//	})
		//})
		//
		//Context("already have pod", func() {
		//	It("Should do nothing", func() {
		//		pod, err := controllerReconciler.generatePod(*mystate, 0)
		//		Expect(err).To(BeNil())
		//
		//		err = k8sClient.Create(ctx, &pod)
		//		Expect(err).To(BeNil())
		//
		//		controllerReconciler := &MyStateReconciler{
		//			Client: k8sClient,
		//			Scheme: k8sClient.Scheme(),
		//		}
		//
		//		newState := mystate.DeepCopy()
		//		controllerReconciler.addFinalizer(newState)
		//		err = k8sClient.Update(ctx, newState)
		//		Expect(err).To(BeNil())
		//
		//		_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
		//			NamespacedName: typeNamespacedName,
		//		})
		//		Expect(err).To(BeNil())
		//	})
		//})
		//
		//Context("upgrade new version", func() {
		//	It("Should remove pod", func() {
		//		pod, err := controllerReconciler.generatePod(*mystate, 0)
		//		Expect(err).To(BeNil())
		//
		//		err = k8sClient.Create(ctx, &pod)
		//		Expect(err).To(BeNil())
		//
		//		controllerReconciler := &MyStateReconciler{
		//			Client: k8sClient,
		//			Scheme: k8sClient.Scheme(),
		//		}
		//
		//		newState := mystate.DeepCopy()
		//		newState.Generation = 2
		//
		//		err = k8sClient.Update(ctx, newState)
		//		Expect(err).To(BeNil())
		//
		//		_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
		//			NamespacedName: typeNamespacedName,
		//		})
		//		Expect(err).To(BeNil())
		//
		//		var childPods v1.PodList
		//
		//		err = k8sClient.List(ctx, &childPods, client.InNamespace(newState.Namespace), client.MatchingFields{podOwnerKey: newState.Name})
		//		Expect(err).To(BeNil())
		//		Expect(len(childPods.Items)).To(Equal(0))
		//	})
		//})
		//It("should return nil", func() {
		//		//	err := k8sClient.Delete(ctx, mystate)
		//		//	Expect(err).To(BeNil())
		//		//	_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
		//		//		NamespacedName: typeNamespacedName,
		//		//	})
		//		//	Expect(err).To(BeNil())
		//		//	k8sClient.Get(ctx, typeNamespacedName, mystate)
		//		//	Expect(len(mystate.GetFinalizers())).To(Equal(1))
		//		//})
		//		//
	})
})
