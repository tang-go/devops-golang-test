package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

func emptyReturn() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func errorReturn(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

func retryReturn() (ctrl.Result, error) {
	return ctrl.Result{Requeue: true}, nil
}

func retryAfterReturn(duration time.Duration) (ctrl.Result, error) {
	return ctrl.Result{Requeue: true, RequeueAfter: duration}, nil
}
