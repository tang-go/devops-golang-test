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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	myappv1 "hzy.com/mystate/api/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

func retry(timeout, interval time.Duration, f func() (bool, error)) error {
	start := time.Now()
	for time.Now().Sub(start) < timeout {
		success, err := f()
		if err != nil {
			return err
		}
		if success {
			return nil
		}
		time.Sleep(interval)
	}
	return fmt.Errorf("timed out waiting for %v to complete", timeout)
}

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
				Replicas: 2,
				Selector: &metav1.LabelSelector{
					MatchLabels:      nil,
					MatchExpressions: nil,
				},
				Template: myappv1.PodTemplateSpec{
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
				VolumeClaimTemplates: []myappv1.PVCTemplate{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "test"},
						Spec: v1.PersistentVolumeClaimSpec{
							VolumeName: "test-pvc",
						},
					},
				},
				ServiceName: "",
				Ordinals:    myappv1.MyOrdinals{},
			},
		}
		ctx = context.Background()
	})

	Context("When reconciling a resource", func() {
		It("Should update", func() {
			var err error
			tryGet := &myappv1.MyState{}
			Expect(k8sClient.Create(ctx, mystate)).To(Succeed())

			// create mystate
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, typeNamespacedName, tryGet)).To(Succeed())
			}, timeout, interval).Should(Succeed())

			// add finalizer
			Eventually(func(g Gomega) {
				k8sClient.Get(ctx, typeNamespacedName, tryGet)
				g.Expect(len(tryGet.GetFinalizers())).To(Equal(1))
			}, timeout, interval).Should(Succeed())

			// create pods
			// wait for pod to ready
			Eventually(func(g Gomega) {
				var childPods v1.PodList
				err := k8sClient.List(ctx, &childPods)
				g.Expect(err).To(BeNil())
				By("Set pod to ready")
				for _, pod := range childPods.Items {
					pod.Status.Conditions = append(pod.Status.Conditions, v1.PodCondition{Type: v1.PodReady, Status: v1.ConditionTrue})
					g.Expect(k8sClient.Status().Update(ctx, &pod)).To(Succeed())
				}

				g.Expect(k8sClient.Get(ctx, typeNamespacedName, tryGet)).To(Succeed())
				g.Expect(tryGet.Status.Replicas).To(Equal(2))
			}, timeout, interval).Should(Succeed())

			// upgrade mystate
			err = retry(timeout, interval, func() (bool, error) {
				Expect(k8sClient.Get(ctx, typeNamespacedName, tryGet)).To(Succeed())
				tryGet.Spec.Template.Spec.Containers[0].Env = []v1.EnvVar{{
					Name:  "k",
					Value: "v",
				}}
				err := k8sClient.Update(ctx, tryGet)
				if err != nil {
					return false, nil
				}
				return true, nil
			})
			Expect(err).To(BeNil())

			Eventually(func(g Gomega) {
				var childPods v1.PodList
				err := k8sClient.List(ctx, &childPods)
				g.Expect(err).To(BeNil())

				By("Set pod to ready")
				for _, pod := range childPods.Items {
					pod.Status.Conditions = append(pod.Status.Conditions, v1.PodCondition{Type: v1.PodReady, Status: v1.ConditionTrue})
					g.Expect(k8sClient.Status().Update(ctx, &pod)).To(Succeed())
				}
				g.Expect(len(childPods.Items[0].Spec.Containers[0].Env)).To(Equal(1))
				g.Expect(k8sClient.Get(ctx, typeNamespacedName, tryGet)).To(Succeed())
				g.Expect(tryGet.Status.Replicas).To(Equal(2))
				g.Expect(len(childPods.Items)).To(Equal(2))
			}, timeout, interval).Should(Succeed())

			// add pods
			By("raise replicas")
			//var tryGet *myappv1.MyState
			for i := 0; i < 5; i++ {
				Expect(k8sClient.Get(ctx, typeNamespacedName, tryGet)).To(Succeed())
				tryGet.Spec.Replicas = 3
				if err := k8sClient.Update(ctx, tryGet); err == nil {
					break
				}
			}

			// wait for all pod ready
			Eventually(func(g Gomega) {
				var childPods v1.PodList
				err := k8sClient.List(ctx, &childPods)
				g.Expect(err).To(BeNil())

				for _, pod := range childPods.Items {
					By("Set pod to ready")
					pod.Status.Conditions = append(pod.Status.Conditions, v1.PodCondition{Type: v1.PodReady, Status: v1.ConditionTrue})
					g.Expect(k8sClient.Status().Update(ctx, &pod)).To(Succeed())
				}

				g.Expect(k8sClient.Get(ctx, typeNamespacedName, tryGet)).To(Succeed())
				g.Expect(tryGet.Spec.Replicas).To(Equal(3))
				g.Expect(k8sClient.Get(ctx, typeNamespacedName, tryGet)).To(Succeed())
				g.Expect(tryGet.Status.Replicas).To(Equal(3))
				g.Expect(len(childPods.Items)).To(Equal(3))
			}, timeout, interval).Should(Succeed())

			// delete pod which is more than replicas
			By("reduce replicas")
			for i := 0; i < 5; i++ {
				Expect(k8sClient.Get(ctx, typeNamespacedName, tryGet)).To(Succeed())
				tryGet.Spec.Replicas = 1
				if err = k8sClient.Update(ctx, tryGet); err == nil {
					break
				}
			}
			Expect(err).To(BeNil())
			Eventually(func(g Gomega) {
				var childPods v1.PodList
				err := k8sClient.List(ctx, &childPods)
				for _, pod := range childPods.Items {
					By("Set pod to ready")
					pod.Status.Conditions = append(pod.Status.Conditions, v1.PodCondition{Type: v1.PodReady, Status: v1.ConditionTrue})
					g.Expect(k8sClient.Status().Update(ctx, &pod)).To(Succeed())
				}
				g.Expect(err).To(BeNil())
				Expect(k8sClient.Get(ctx, typeNamespacedName, tryGet)).To(Succeed())
				g.Expect(len(childPods.Items)).To(Equal(1))
				g.Expect(k8sClient.Get(ctx, typeNamespacedName, tryGet)).To(Succeed())
				g.Expect(tryGet.Status.Replicas).To(Equal(1))
			}, timeout, interval).Should(Succeed())

			// delete my state will clear pods created by it
			By("delete mystate")
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			Expect(k8sClient.Delete(ctx, mystate)).To(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, typeNamespacedName, tryGet)).NotTo(Succeed())

				var childPods v1.PodList
				err := k8sClient.List(ctx, &childPods)
				g.Expect(err).To(BeNil())
				g.Expect(len(childPods.Items)).To(Equal(0))
			})
		})

	})
})
