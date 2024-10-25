package controller

import (
	"context"
	myappv1 "hzy.com/mystate/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *MyStateReconciler) isDeleting(state myappv1.MyState) bool {
	return !state.GetDeletionTimestamp().IsZero()
}

func (r *MyStateReconciler) addFinalizer(state *myappv1.MyState) {
	controllerutil.AddFinalizer(state, MyStateFinalizerName)
}

func (r *MyStateReconciler) hasFinalizer(state myappv1.MyState) bool {
	return controllerutil.ContainsFinalizer(&state, MyStateFinalizerName)
}

func (r *MyStateReconciler) delete(ctx context.Context, state myappv1.MyState) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	childs, err := r.orderedChildPods(ctx, state)
	if err != nil {
		return errorReturn(err)
	}

	if len(childs) == 0 {
		log.Info("all child of MyState is deleted")
		controllerutil.RemoveFinalizer(&state, MyStateFinalizerName)
		if err := r.Update(ctx, &state); err != nil {
			return errorReturn(err)
		} else {
			return emptyReturn()
		}
	}

	lastChild := childs[len(childs)-1]
	if !lastChild.DeletionTimestamp.IsZero() {
		log.Info("child of MyState is deleting", "pod", lastChild.Name)
		return retryReturn()
	} else {
		log.Info("start to delete child of MyState", "pod", lastChild.Name)
		return r.removePod(ctx, lastChild)
	}

}
