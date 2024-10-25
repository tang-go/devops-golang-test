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
	myappv1 "hzy.com/mystate/api/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sort"
)

// MyStateReconciler reconciles a MyState object
type MyStateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	jobOwnerKey = ".metadata.controller"
)

// +kubebuilder:rbac:groups=myapp.hzy.com,resources=mystates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=myapp.hzy.com,resources=mystates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=myapp.hzy.com,resources=mystates/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pod,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pod/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyState object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *MyStateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	var myState myappv1.MyState
	if err := r.Get(ctx, req.NamespacedName, &myState); err != nil {
		log.Error(err, "unable to fetch MyState")
		//忽略掉 not-found 错误，它们不能通过重新排队修复（要等待新的通知）
		//在删除一个不存在的对象时，可能会报这个错误。
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if r.isDeleting(myState) {
		if r.hasFinalizer(myState) {
			log.Info("start to delete myState", "myState", myState.Name)
			return r.delete(ctx, myState)
		} else {
			return emptyReturn()
		}
	} else if !r.hasFinalizer(myState) {
		log.Info("add finalizer to myState %s", "myState", myState.Name)
		r.addFinalizer(&myState)
		if err := r.Update(ctx, &myState); err != nil {
			return errorReturn(err)
		} else {
			return retryReturn()
		}
	}

	return r.reconcile(ctx, myState)
}

func (r *MyStateReconciler) reconcile(ctx context.Context, myState myappv1.MyState) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// get all child pods
	childPods, err := r.orderedChildPods(ctx, myState)
	if err != nil {
		return errorReturn(err)
	}
	log.Info("got child pods", "nums", len(childPods))

	if waitFor, needWait := r.needWaitForPod(childPods); needWait {
		log.Info("waiting for pod to be ready", "pod", waitFor.Name)
		return retryReturn()
	}

	// all exist pods are ready
	myState.Status.ReadyReplicas = myState.Status.Replicas
	if err = r.Status().Update(ctx, &myState); err != nil {
		return errorReturn(err)
	}

	// make sure there are enough pods
	if missingPod, exist := getNextMissingPod(myState, childPods); exist {
		log.Info("creating pod", "pod", missingPod.Name)
		myState.Status.Replicas = len(childPods) + 1
		if err = r.Status().Update(ctx, &myState); err != nil {
			return errorReturn(err)
		}
		return r.createPod(ctx, missingPod)
	}

	if needRemovePod, exist := getNeedRemovePod(myState, childPods); exist {
		log.Info("deleting pod", "pod", needRemovePod.Name)
		return r.removePod(ctx, needRemovePod)
	}

	if needUpgrade(myState) {
		return r.upgrade(ctx, myState, childPods)
	}

	return emptyReturn()
}

func (r *MyStateReconciler) orderedChildPods(ctx context.Context, myState myappv1.MyState) ([]v1.Pod, error) {
	var childPods v1.PodList
	if err := r.List(ctx, &childPods, client.InNamespace(myState.Namespace), client.MatchingFields{jobOwnerKey: myState.Name}); err != nil {
		return nil, err
	}
	sort.Slice(childPods.Items, func(i, j int) bool {
		return getIndexOfPod(&childPods.Items[i]) < getIndexOfPod(&childPods.Items[j])
	})
	return childPods.Items, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyStateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&myappv1.MyState{}).
		Named("mystate").
		Complete(r)
}
