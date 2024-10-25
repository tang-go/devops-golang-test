package controller

import (
	"context"
	"fmt"
	myappv1 "hzy.com/mystate/api/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strconv"
	"strings"
)

func (r *MyStateReconciler) needWaitForPod(pods []v1.Pod) (*v1.Pod, bool) {
	for _, pod := range pods {
		if !isPodReady(pod) {
			return &pod, true
		}
	}
	return nil, false
}

func isPodReady(pod v1.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == v1.PodReady && c.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

func (r *MyStateReconciler) createPod(ctx context.Context, pod v1.Pod) (ctrl.Result, error) {
	err := r.Create(ctx, &pod)
	if err != nil {
		return errorReturn(err)
	}
	return retryReturn()
}

func (r *MyStateReconciler) removePod(ctx context.Context, pod v1.Pod) (ctrl.Result, error) {
	err := r.Delete(ctx, &pod)
	if err != nil {
		return errorReturn(err)
	}
	return retryReturn()
}

func (r *MyStateReconciler) updatePod(ctx context.Context, pod v1.Pod) (ctrl.Result, error) {
	err := r.Update(ctx, &pod)
	if err != nil {
		return errorReturn(err)
	}
	return reconcile.Result{}, nil
}

func getNextMissingPod(myState myappv1.MyState, pods []v1.Pod) (v1.Pod, bool) {
	podMap := make(map[string]v1.Pod)
	for _, pod := range pods {
		podMap[pod.Name] = pod
	}

	for i := myState.Spec.Ordinals.Start; i < myState.Spec.Ordinals.Start+myState.Spec.Replicas; i++ {
		if _, exist := podMap[podNameFor(myState, i)]; !exist {
			return generatePod(myState, i), true
		}
	}
	return v1.Pod{}, false
}

func getNeedRemovePod(myState myappv1.MyState, pods []v1.Pod) (v1.Pod, bool) {
	if len(pods) == 0 {
		return v1.Pod{}, false
	}
	lastPod := pods[len(pods)-1]
	if getIndexOfPod(&lastPod) >= myState.Spec.Replicas {
		return lastPod, true
	}
	return v1.Pod{}, false
}

func podNameFor(myState myappv1.MyState, index int) string {
	return fmt.Sprintf("%s-%d", myState.Name, index)
}

func generatePod(myState myappv1.MyState, index int) v1.Pod {
	// create pod object
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        podNameFor(myState, index),
			Namespace:   myState.Namespace,
			Labels:      myState.Spec.Template.Labels,
			Annotations: myState.Spec.Template.Annotations,
			Finalizers:  myState.Spec.Template.Finalizers,
		},
		Spec: myState.Spec.Template.Spec,
	}

	// about network
	pod.Spec.Hostname = pod.Name
	pod.Spec.Subdomain = myState.Spec.ServiceName

	// add labels
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels[PodNameLabelName] = pod.Name
	pod.Labels[PodIndexLabelName] = strconv.Itoa(index)
	pod.Labels[MyStateGenerationLabelName] = fmt.Sprintf("%d", myState.GetGeneration())

	addStorageToPod(myState, &pod)
	return pod
}

func addStorageToPod(myState myappv1.MyState, pod *v1.Pod) {
	currentVolumes := pod.Spec.Volumes
	pvcMaps := getPvcForPod(&myState, pod)
	newVolumes := make([]v1.Volume, 0, len(pvcMaps))
	for name, claim := range pvcMaps {
		newVolumes = append(newVolumes, v1.Volume{
			Name: name,
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: claim.Name,
				},
			},
		})
	}

	// add volumes if not exist
	for i := range currentVolumes {
		if _, ok := pvcMaps[currentVolumes[i].Name]; !ok {
			newVolumes = append(newVolumes, currentVolumes[i])
		}
	}
	pod.Spec.Volumes = newVolumes
}

func getPvcForPod(myState *myappv1.MyState, pod *v1.Pod) map[string]v1.PersistentVolumeClaim {
	index := getIndexOfPod(pod)
	templates := myState.Spec.VolumeClaimTemplates
	pvcs := make(map[string]v1.PersistentVolumeClaim, len(templates))
	for i := range templates {
		pvc := templates[i].DeepCopy()
		pvc.Name = pvcNameFor(myState, pvc, index)
		pvc.Namespace = myState.Namespace
		if pvc.Labels != nil {
			for key, value := range myState.Spec.Selector.MatchLabels {
				pvc.Labels[key] = value
			}
		} else {
			pvc.Labels = myState.Spec.Selector.MatchLabels
		}
		pvcs[templates[i].Name] = *pvc
	}
	return pvcs
}

func getIndexOfPod(pod *v1.Pod) int {
	strs := strings.Split(pod.Name, "-")
	index, err := strconv.Atoi(strs[len(strs)-1])
	if err != nil {
		return -1
	}
	return index
}

func pvcNameFor(myState *myappv1.MyState, pvc *v1.PersistentVolumeClaim, index int) string {
	return fmt.Sprintf("%s-%s-%d", pvc.Name, myState.Name, index)
}
