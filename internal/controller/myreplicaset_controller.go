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
	"math"
	"math/rand"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mygroupv1 "my.domain/go-test/api/v1"
)

const (
	typeAvailableReplicasets = "Available"
)

// MyReplicaSetReconciler reconciles a MyReplicaSet object
type MyReplicaSetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// +kubebuilder:rbac:groups=mygroup.my.domain,resources=myreplicasets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mygroup.my.domain,resources=myreplicasets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mygroup.my.domain,resources=myreplicasets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyReplicaSet object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *MyReplicaSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	mRS := &mygroupv1.MyReplicaSet{}
	if err := r.Get(ctx, req.NamespacedName, mRS); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if mRS.Status.Conditions == nil || len(mRS.Status.Conditions) == 0 {
		meta.SetStatusCondition(&mRS.Status.Conditions, metav1.Condition{
			Type:    typeAvailableReplicasets,
			Status:  metav1.ConditionUnknown,
			Reason:  "Reconciling",
			Message: "Starting reconciliation",
		})

		if err := r.Status().Update(ctx, mRS); err != nil {
			return ctrl.Result{}, err
		}

	}

	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(req.Namespace),
		client.MatchingLabels(mRS.Spec.Template.ObjectMeta.Labels),
	}
	if err := r.List(ctx, podList, listOpts...); err != nil {
		return ctrl.Result{}, err
	}

	podNames := getPodNames(podList.Items)
	currentCount := len(podList.Items)
	specCount := mRS.Spec.Replicas
	countDiff := int(math.Abs(float64(currentCount - int(specCount))))
	if currentCount < int(specCount) {
		for i := 0; i < countDiff; i++ {
			pod, err := r.createPodTemplate(mRS)
			if err != nil {
				meta.SetStatusCondition(&mRS.Status.Conditions, metav1.Condition{
					Type:    typeAvailableReplicasets,
					Status:  metav1.ConditionFalse,
					Reason:  "Reconciling",
					Message: fmt.Sprintf("Failed to define new pod for the custom resource (%s): (%s)", mRS.Name, err),
				})
				if err := r.Status().Update(ctx, mRS); err != nil {
					return ctrl.Result{}, nil
				}
				return ctrl.Result{}, nil
			}
			if err := r.Create(ctx, pod); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{Requeue: true}, nil
	} else if currentCount > int(specCount) {
		for i := 0; i < countDiff; i++ {
			pod := &corev1.Pod{}
			deletePodName := podNames[i]
			if err := r.Get(ctx, types.NamespacedName{Name: deletePodName, Namespace: mRS.Namespace}, pod); err != nil {
				meta.SetStatusCondition(&mRS.Status.Conditions, metav1.Condition{
					Type:    typeAvailableReplicasets,
					Status:  metav1.ConditionFalse,
					Reason:  "Reconciling",
					Message: fmt.Sprintf("Failed to fetch the pod to be deleted for the custom resource (%s): (%s)", mRS.Name, err),
				})
				if err = r.Status().Update(ctx, mRS); err != nil {
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, err
			}
			if err := r.Delete(ctx, pod, client.GracePeriodSeconds(30)); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{Requeue: true}, nil
	}

	meta.SetStatusCondition(&mRS.Status.Conditions, metav1.Condition{
		Type:    typeAvailableReplicasets,
		Status:  metav1.ConditionTrue,
		Reason:  "Reconciling",
		Message: fmt.Sprintf("Pods for custom resource (%s) reconcile successfully", mRS.Name),
	})

	mRS.Status.Pods = podNames
	if err := r.Status().Update(ctx, mRS); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyReplicaSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mygroupv1.MyReplicaSet{}).
		Complete(r)
}

func generatePodName() string {
	letters := []byte("1234567890abcdefghijklmnopqrstuvwxyz")
	ranStr := make([]byte, 5)
	for i := 0; i < 5; i++ {
		ranStr[i] = letters[rand.Intn(len(letters))]
	}
	str := string(ranStr)
	return str
}

func getPodNames(podList []corev1.Pod) []string {
	podNames := []string{}
	for _, pod := range podList {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func (r *MyReplicaSetReconciler) createPodTemplate(mRS *mygroupv1.MyReplicaSet) (*corev1.Pod, error) {
	template := mRS.Spec.Template
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%v-%v", mRS.Name, generatePodName()),
			Namespace: mRS.Namespace,
			Labels:    template.ObjectMeta.Labels,
		},
		Spec: template.Spec,
	}

	if err := ctrl.SetControllerReference(mRS, pod, r.Scheme); err != nil {
		return nil, err
	}
	return pod, nil
}
