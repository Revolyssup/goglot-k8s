/*
Copyright 2021.

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

package controllers

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	goglotdevv1alpha1 "revolyssup/goglot-k8s/api/v1alpha1"

	batchv1 "k8s.io/api/batch/v1"
)

// GlotpodReconciler reconciles a Glotpod object
type GlotpodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=goglot.dev.github.com,resources=glotpods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=goglot.dev.github.com,resources=glotpods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=goglot.dev.github.com,resources=glotpods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Glotpod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *GlotpodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// your logic here
	pod := &goglotdevv1alpha1.Glotpod{}
	err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, pod)
	if err != nil {
		return ctrl.Result{}, nil
	}
	err = r.createJob(ctx, pod)
	if err != nil {
		fmt.Println("loldu ", err.Error())
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GlotpodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&goglotdevv1alpha1.Glotpod{}).
		Complete(r)
}

var DEFAULTCOUNT int32 = 1

const (
	jsscript = "./runjs.sh"
)

func (r *GlotpodReconciler) createJob(ctx context.Context, pod *goglotdevv1alpha1.Glotpod) error {
	myJob := batchv1.Job{}
	fmt.Println("YE MILA ", pod)
	var script string
	switch pod.Spec.Language {
	case "js":
		script = jsscript
	}
	fmt.Println([]string{script, pod.Spec.Input, pod.Spec.Code, "test.js"})
	var maxPodCount int32 = 1
	var maxTimeToCompletion int64 = 2
	var deleteAfterSeconds int32 = 2
	myJob = batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: pod.Namespace,
			Name:      pod.Name,
		},
		Spec: batchv1.JobSpec{
			Parallelism:             &maxPodCount,
			Completions:             &maxPodCount,
			ActiveDeadlineSeconds:   &maxTimeToCompletion,
			TTLSecondsAfterFinished: &deleteAfterSeconds,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: pod.Namespace,
					Labels: map[string]string{
						"name": pod.Name,
					},
				},
				Spec: v1.PodSpec{

					Containers: []v1.Container{
						{
							Name:    "testcontainer",
							Image:   "revoly/jsrunnerv2:oggg",
							Command: []string{script, pod.Spec.Input, pod.Spec.Code, "test.js"},
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}
	return r.Create(ctx, &myJob)
}
