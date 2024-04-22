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

package controllers

import (
	"context"
	"os"
	"strconv"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/jsm84/multus-notifier-operator/api/v1alpha1"
)

// MultusNotifierReconciler reconciles a MultusNotifier object
type MultusNotifierReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.f-i.de,resources=multusnotifiers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.f-i.de,resources=multusnotifiers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.f-i.de,resources=multusnotifiers/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps.f-i.de,resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MultusNotifier object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *MultusNotifierReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MultusNotifier")

	// Fetch the MultusNotifer instance
	instance := &noopv1alpha1.MultusNotifier{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Pod object
	pods := newPodsForCR(instance)

	// Check if each Pod already exists
	found := []corev1.Pod{}

	for _, pod := range pods {
		// Set MultusNotifier instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		err = r.client.List(context.TODO(), labels.Selector{Label: instance.PodLabel}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Getting List of Pods", "labels.Selector", instance.PodLabel)
			err = r.client.Create(context.TODO(), pod)
			if err != nil {
				return reconcile.Result{}, err
			}

			// Pod created successfully - don't requeue
			return reconcile.Result{}, nil
		} else if err != nil {
			return reconcile.Result{}, err
		}

		// Pod already exists - don't requeue
		reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	}

	// Update CR Status once deployed
	if instance.Status.Deployed != true {
		instance.Status.Deployed = true
		reqLogger.Info("Updating Object Status", "CR.Kind", instance.Kind, "CR.Name", instance.Name)
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update pod deploy status")
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

// newPodsForCR returns a ubi-minimal pod with the same name/namespace as the cr
func newPodsForCR(cr *noopv1alpha1.UBINoOp) []*corev1.Pod {
	// get the container image location from the expected env var
	// default to ubi8 minimal image from registry.redhat.io
	ubiImg := os.Getenv("RELATED_IMAGE_UBI_MINIMAL")
	if ubiImg == "" {
		ubiImg = "registry.redhat.io/ubi8/ubi-minimal:latest"
	}

	// podLst is a slice of pod whose len matches cr.Spec.Size
	podLst := make([]*corev1.Pod, cr.Spec.Size)

	labels := map[string]string{
		"app": cr.Name,
	}

	for idx, _ := range podLst {
		i := strconv.Itoa(idx)
		podLst[idx] = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cr.Name + "-pod" + i,
				Namespace: cr.Namespace,
				Labels:    labels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "ubi-minimal",
						Image:   ubiImg,
						Command: []string{"sleep", "3600"},
					},
				},
			},
		}
	}

	return podLst
}

// SetupWithManager sets up the controller with the Manager.
func (r *MultusNotifierReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.MultusNotifier{}).
		Complete(r)
}
