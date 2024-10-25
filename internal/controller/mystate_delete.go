package controller

import (
	"context"
	myappv1 "hzy.com/mystate/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *MyStateReconciler) isDeleting(state myappv1.MyState) bool {
	return !state.GetDeletionTimestamp().IsZero()
}

func (r *MyStateReconciler) addFinalizer(state myappv1.MyState) {
	controllerutil.AddFinalizer(&state, MyStateFinalizerName)
}

func (r *MyStateReconciler) hasFinalizer(state myappv1.MyState) bool {
	return controllerutil.ContainsFinalizer(&state, MyStateFinalizerName)
}

func (r *MyStateReconciler) delete(ctx context.Context, state myappv1.MyState) (ctrl.Result, error) {

}
