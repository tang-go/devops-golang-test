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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	myrsv1 "github.com/wholj/myreplicaset/api/v1"
)

// MyReplicaSetReconciler reconciles a MyReplicaSet object
type MyReplicaSetReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder // 记录事件
}

// +kubebuilder:rbac:groups=myrs.github.com,resources=myreplicasets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=myrs.github.com,resources=myreplicasets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=myrs.github.com,resources=myreplicasets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyReplicaSet object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *MyReplicaSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	myReplicaSet := &myrsv1.MyReplicaSet{}

	// Fetch the MyReplicaSet instance
	if err := r.Get(ctx, req.NamespacedName, myReplicaSet); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 获取期望的副本数
	desireReplicas := myReplicaSet.Spec.Replicas
	fmt.Printf("===> desireReplicas: %v\n", desireReplicas)

	// 获取当前副本数
	currentPods := &corev1.PodList{}
	labelSelector := client.MatchingLabels{"app": myReplicaSet.Name}
	fmt.Printf("---> labelSelector: %v\n", labelSelector)
	if err := r.Client.List(ctx, currentPods, labelSelector); err != nil {
		return ctrl.Result{}, err
	}
	// 更新 .Status.replicas
	// currentReplicaCount := int32(len(currentPods.Items))
	currentReplicaCount := int32(0)
	for _, pod := range currentPods.Items {
		if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodPending {
			currentReplicaCount++
		}
	}

	fmt.Printf("<=== currentReplicaCount: %v\n", currentReplicaCount)
	myReplicaSet.Status.Replicas = currentReplicaCount

	// 创建或删除 Pods
	if currentReplicaCount < desireReplicas {
		for i := currentReplicaCount; i < desireReplicas; i++ {
			pod := r.newPodForCR(myReplicaSet, i)
			if err := r.Client.Create(ctx, pod); err != nil {
				// 记录事件
				r.Recorder.Event(myReplicaSet, corev1.EventTypeWarning, "FailedCreatingPod", err.Error())
				return ctrl.Result{}, err
			}
			// 记录事件
			r.Recorder.Event(myReplicaSet, corev1.EventTypeNormal, "CreatedPod", pod.Name)
			klog.Infof("Created Pod %s", pod.Name)
		}
	} else if currentReplicaCount > desireReplicas {
		// 删除多余的 Pod
		for i := currentReplicaCount; i > desireReplicas; i-- {
			pod := currentPods.Items[i-1]
			if err := r.Client.Delete(ctx, &pod); err != nil {
				// 记录事件
				r.Recorder.Event(myReplicaSet, corev1.EventTypeWarning, "FailedDeletePod", err.Error())
				return ctrl.Result{}, err
			}
			// 记录事件
			r.Recorder.Event(myReplicaSet, corev1.EventTypeNormal, "DeletedPod", pod.Name)
			klog.Infof("Delete Pod %s", pod.Name)
		}
	}

	// 更新 .Status 字段
	err := r.Client.Status().Update(ctx, myReplicaSet)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *MyReplicaSetReconciler) newPodForCR(cr *myrsv1.MyReplicaSet, index int32) *corev1.Pod {
	labels := map[string]string{"app": cr.Name}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", cr.Name, index),
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: cr.Spec.Template.Spec,
	}
	return pod
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyReplicaSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("myreplicaset-controller")

	return ctrl.NewControllerManagedBy(mgr).
		For(&myrsv1.MyReplicaSet{}).
		Owns(&corev1.Pod{}). // 监视 Pod 的变化
		Complete(r)
}

// Todo
// 1，创建 Pod -> pkg/controller/replicaset/replica_set.go#L384
// 2，删除 Pod -> pkg/controller/replicaset/replica_set.go#L501
