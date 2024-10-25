package controller

import (
	"context"
	"fmt"
	myappv1 "hzy.com/mystate/api/v1"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func needUpgrade(myState myappv1.MyState) bool {
	return fmt.Sprintf("%d", myState.Generation) != myState.Status.CurrentGeneration
}

func (r *MyStateReconciler) upgrade(ctx context.Context, myState myappv1.MyState, pods []v1.Pod) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	targetGeneration := fmt.Sprintf("%d", myState.Generation)
	log.Info("upgrading MyState generation",
		"MyState", myState.Name,
		"current", myState.Status.CurrentGeneration,
		"target", targetGeneration)
	for i, pod := range pods {
		if pod.Labels[MyStateGenerationLabelName] == targetGeneration {
			if myState.Status.CurrentReplicas < i {
				myState.Status.CurrentReplicas = i
				if err := r.Status().Update(ctx, &myState); err != nil {
					return errorReturn(err)
				} else {
					return retryReturn()
				}
			}
		} else {
			if _, err := r.updatePod(ctx, generatePod(myState, i)); err != nil {
				return errorReturn(err)
			}
		}
	}

	myState.Status.CurrentGeneration = targetGeneration
	if err := r.Status().Update(ctx, &myState); err != nil {
		return errorReturn(err)
	} else {
		return emptyReturn()
	}
}
